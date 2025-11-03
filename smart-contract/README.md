# TLC Smart Contract

Smart contract untuk sistem payment TLC (Telkom Coin) yang terintegrasi dengan Ganache GUI.

## 📋 Features

- **PaymentToken**: ERC20-compatible token untuk sistem payment
- **Ganache GUI Integration**: Deploy dan testing menggunakan Ganache GUI
- **Web3.js**: JavaScript library untuk interaksi dengan blockchain
- **Solidity 0.8.0**: Compatible dengan Ganache GUI

## 🚀 Quick Start

### Prerequisites

1. **Ganache GUI** running di port 8545
2. **Node.js** dan npm terinstall

### Installation

```bash
npm install
```

### Deployment

1. **Start Ganache GUI** di port 8545
2. **Deploy smart contract:**

```bash
npm run deploy
```

atau

```bash
node scripts/deploy.js
```

## 📁 Project Structure

```
smart-contract/
├── contracts/
│   └── PaymentToken.sol      # Main smart contract
├── scripts/
│   └── deploy.js            # Deployment script
├── deployment-info.json     # Deployment information
├── package.json
└── README.md
```

## 🔧 Configuration

Contract akan di-deploy menggunakan:
- **Account 0** dari Ganache GUI
- **Gas Limit**: 12M
- **Gas Price**: 20 gwei
- **Initial Supply**: 1000 TELKOM tokens

## 🔗 Integration

Setelah deployment, update backend `.env` file dengan:
- `CONTRACT_ADDRESS`: Address dari deployed contract
- `ADMIN_WALLET_ADDRESS`: Account 0 dari Ganache GUI
- `ADMIN_PRIVATE_KEY`: Private key dari Account 0

## 🔄 Restart Workflow

Setiap restart laptop:

1. **Start Ganache GUI** (port 8545)
2. **Deploy smart contract:**
   ```bash
   npm run deploy
   ```
3. **Update backend .env** dengan contract address baru
4. **Start backend:**
   ```bash
   cd ../telkom_coin_back_end
   go run cmd/main.go
   ```

## 📝 Notes

- Contract menggunakan Solidity 0.8.0 untuk compatibility dengan Ganache GUI
- EVM version: istanbul (compatible dengan Ganache GUI lama)
- Optimizer disabled untuk menghindari "invalid opcode" error
- Account dan private key konsisten setiap deployment
