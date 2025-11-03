import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const PaymentTokenModule = buildModule("PaymentTokenModule", (m) => {
  // Initial supply: 1,000,000 TELKOM tokens
  // Akan di-mint ke address yang deploy contract
 const initialSupply = 1000000n;;

  // Deploy PaymentToken contract
  const paymentToken = m.contract("PaymentToken", [initialSupply]);

  return { paymentToken };
});

export default PaymentTokenModule;