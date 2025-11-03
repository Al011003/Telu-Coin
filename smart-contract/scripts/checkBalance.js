import { ethers } from "ethers";
import fs from "fs";

async function main() {
  const contractAddress = "0xEe29aC868A3503d571942f7f9Ad7280F91998e41";
  const provider = new ethers.JsonRpcProvider("http://127.0.0.1:8545");
  
  // Gunakan wallet index 0 (yang punya private key itu)
  const wallet = new ethers.Wallet(
    "0xec23f1632afda2f4c219b2cb3de352d9a9dbe3d2b90baf27e2eb884a2bcad4a7",
    provider
  );

  console.log("Caller address:", wallet.address);

  const contractJson = JSON.parse(
    fs.readFileSync("./artifacts/contracts/PaymentToken.sol/PaymentToken.json", "utf8")
  );
  const contract = new ethers.Contract(contractAddress, contractJson.abi, wallet);

  // Cek current balance
  const currentBalance = await contract.balanceOf(wallet.address);
  console.log("Current balance:", ethers.formatEther(currentBalance), "TELKOM");

  // Cek minimum amount
  const minMintAmount = await contract.minMintAmount();
  console.log("Min mint amount:", ethers.formatEther(minMintAmount), "TELKOM");

  // Topup
  const amount = ethers.parseEther("50000"); // 50,000 TELKOM
  console.log("\nTopuping:", ethers.formatEther(amount), "TELKOM");

  try {
    const tx = await contract.instantTopup(amount, "proof123");
    console.log("TX hash:", tx.hash);
    
    const receipt = await tx.wait();
    console.log("✅ Topup success!");
    console.log("Block:", receipt.blockNumber);

    // Cek balance baru
    const newBalance = await contract.balanceOf(wallet.address);
    console.log("New balance:", ethers.formatEther(newBalance), "TELKOM");
  } catch (error) {
    console.error("❌ Error:", error.message);
  }
}

main().catch(console.error);