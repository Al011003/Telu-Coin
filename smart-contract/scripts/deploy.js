import Web3 from 'web3';
import solc from 'solc';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function main() {
  console.log("🚀 Deploying PaymentToken contract to Ganache GUI...");

  // Connect to Ganache GUI
  const web3 = new Web3('http://127.0.0.1:8545');
  
  // Check connection
  try {
    const blockNumber = await web3.eth.getBlockNumber();
    console.log("✅ Connected to Ganache GUI, current block:", blockNumber);
  } catch (error) {
    console.error("❌ Failed to connect to Ganache GUI:", error.message);
    console.log("💡 Make sure Ganache GUI is running on port 8545");
    process.exit(1);
  }

  // Get accounts from Ganache GUI
  const accounts = await web3.eth.getAccounts();
  const deployer = accounts[0];
  console.log("Deploying with account:", deployer);
  
  const balance = await web3.eth.getBalance(deployer);
  console.log("Account balance:", web3.utils.fromWei(balance, 'ether'), "ETH");

  // Read and compile contract
  const contractPath = path.join(__dirname, '../contracts/PaymentToken.sol');
  const openzeppelinPath = path.join(__dirname, '../node_modules/@openzeppelin/contracts');
  
  if (!fs.existsSync(contractPath)) {
    console.error("❌ PaymentToken.sol not found");
    process.exit(1);
  }

  const source = fs.readFileSync(contractPath, 'utf8');
  
  // Solidity compiler input
  const input = {
    language: 'Solidity',
    sources: {
      'PaymentToken.sol': {
        content: source
      }
    },
    settings: {
      outputSelection: {
        '*': {
          '*': ['abi', 'evm.bytecode']
        }
      },
      evmVersion: 'istanbul', // Compatible dengan Ganache GUI
      optimizer: {
        enabled: false // Disable optimizer untuk compatibility
      }
    }
  };

  // Import callback untuk OpenZeppelin
  function findImports(importPath) {
    try {
      if (importPath.startsWith('@openzeppelin/')) {
        const fullPath = path.join(openzeppelinPath, importPath.replace('@openzeppelin/contracts/', ''));
        if (fs.existsSync(fullPath)) {
          return { contents: fs.readFileSync(fullPath, 'utf8') };
        }
      }
      return { error: 'File not found' };
    } catch (error) {
      return { error: 'Import error: ' + error.message };
    }
  }

  console.log("📦 Compiling contract...");
  
  try {
    const output = JSON.parse(solc.compile(JSON.stringify(input), { import: findImports }));
    
    if (output.errors) {
      const hasErrors = output.errors.some(error => error.severity === 'error');
      if (hasErrors) {
        console.error("❌ Compilation errors:");
        output.errors.forEach(error => {
          if (error.severity === 'error') {
            console.error(error.formattedMessage);
          }
        });
        process.exit(1);
      } else {
        console.log("⚠️  Compilation warnings (non-critical):");
        output.errors.forEach(error => {
          if (error.severity === 'warning') {
            console.log(error.formattedMessage);
          }
        });
      }
    }

    const contract = output.contracts['PaymentToken.sol']['PaymentToken'];
    const abi = contract.abi;
    const bytecode = contract.evm.bytecode.object;

    console.log("✅ Contract compiled successfully");

    // Create contract instance
    const contractInstance = new web3.eth.Contract(abi);
    
    // Initial supply: 1000 tokens (constructor akan multiply dengan 10^18)
    // Jadi total supply = 1000 * 10^18 = 1000 TELKOM tokens
    const initialSupply = "1000";
    
    console.log("🚀 Deploying contract with initial supply:", initialSupply, "tokens");
    
    // Deploy contract
    const deployTx = contractInstance.deploy({
      data: '0x' + bytecode,
      arguments: [initialSupply]
    });

    // Estimate gas
    const gas = await deployTx.estimateGas({ from: deployer });
    console.log("⛽ Estimated gas:", gas.toString());

    // Send deployment transaction
    const deployedContract = await deployTx.send({
      from: deployer,
      gas: Math.floor(Number(gas) * 1.2), // Add 20% buffer
      gasPrice: '20000000000' // 20 gwei
    });

    console.log("\n🎉 PaymentToken deployed successfully!");
    console.log("📍 Contract Address:", deployedContract.options.address);
    console.log("👤 Owner Address:", deployer);
    console.log("🔗 Transaction Hash:", deployedContract.transactionHash);
    
    // Verify deployment
    const name = await deployedContract.methods.name().call();
    const symbol = await deployedContract.methods.symbol().call();
    const totalSupply = await deployedContract.methods.totalSupply().call();
    const ownerBalance = await deployedContract.methods.balanceOf(deployer).call();
    
    console.log("\n🔍 Contract Verification:");
    console.log("Name:", name);
    console.log("Symbol:", symbol);
    console.log("Total Supply:", web3.utils.fromWei(totalSupply, 'ether'), "TELKOM");
    console.log("Owner Balance:", web3.utils.fromWei(ownerBalance, 'ether'), "TELKOM");
    
    // Save deployment info untuk backend
    const blockNumber = await web3.eth.getBlockNumber();
    const deploymentInfo = {
      contractAddress: deployedContract.options.address,
      ownerAddress: deployer,
      transactionHash: deployedContract.transactionHash,
      blockNumber: blockNumber.toString(),
      timestamp: new Date().toISOString(),
      network: "ganache-gui",
      chainId: 1337,
      gasUsed: gas.toString()
    };
    
    const deploymentPath = path.join(__dirname, '../deployment-info.json');
    fs.writeFileSync(deploymentPath, JSON.stringify(deploymentInfo, null, 2));
    console.log("\n💾 Deployment info saved to:", deploymentPath);
    
    console.log("\n📝 UPDATE BACKEND .env FILE:");
    console.log("CONTRACT_ADDRESS=" + deployedContract.options.address);
    console.log("ADMIN_WALLET_ADDRESS=" + deployer);
    console.log("ADMIN_PRIVATE_KEY=ec23f1632afda2f4c219b2cb3de352d9a9dbe3d2b90baf27e2eb884a2bcad4a7");
    
    return deployedContract.options.address;
    
  } catch (error) {
    console.error("❌ Deployment failed:", error.message);
    if (error.message.includes('revert')) {
      console.log("💡 Contract reverted - check constructor parameters");
    }
    process.exit(1);
  }
}

main()
  .then((address) => {
    console.log("\n🎉 Deployment completed successfully!");
    console.log("🔗 Contract address:", address);
    console.log("🌐 You can now see the transaction in Ganache GUI!");
    process.exit(0);
  })
  .catch((error) => {
    console.error("❌ Script failed:", error);
    process.exit(1);
  });
