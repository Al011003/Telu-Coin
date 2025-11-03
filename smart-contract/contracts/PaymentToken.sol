// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title PaymentToken
 * @dev Fully Decentralized ERC20 Token untuk sistem payment blockchain
 * Token TELKOM dapat di-mint oleh siapa saja dengan proof of payment (IDR)
 * dan di-burn untuk withdraw ke IDR tanpa admin approval
 */
contract PaymentToken is ERC20 {
    // Decimals default adalah 18 (seperti ETH)
    uint8 private _decimals = 18;

    // Exchange rate: 1 IDR = 1 TELKOM token (bisa diubah via governance)
    uint256 public exchangeRate = 1e18; // 1 token = 1 IDR (dalam wei)

    // Minimum amounts untuk prevent spam
    uint256 public minMintAmount = 10000 * 1e18; // 10,000 TELKOM minimum
    uint256 public minBurnAmount = 1000 * 1e18;  // 1,000 TELKOM minimum

    // Struct untuk top-up requests
    struct TopupRequest {
        address user;
        uint256 amount;
        string paymentProof; // IPFS hash atau URL bukti transfer
        uint256 timestamp;
        bool processed;
    }

    // Struct untuk withdraw requests
    struct WithdrawRequest {
        address user;
        uint256 amount;
        string bankAccount; // Rekening tujuan
        uint256 timestamp;
        bool processed;
    }

    // Storage
    mapping(bytes32 => TopupRequest) public topupRequests;
    mapping(bytes32 => WithdrawRequest) public withdrawRequests;
    mapping(address => uint256) public userNonces; // Prevent replay attacks

    // Events untuk tracking
    event TokensMinted(address indexed to, uint256 amount, bytes32 indexed requestId);
    event TokensBurned(address indexed from, uint256 amount, bytes32 indexed requestId);
    event PaymentProcessed(address indexed from, address indexed to, uint256 amount, string note);
    event TopupRequested(address indexed user, uint256 amount, bytes32 indexed requestId, string paymentProof);
    event WithdrawRequested(address indexed user, uint256 amount, bytes32 indexed requestId, string bankAccount);
    event ExchangeRateUpdated(uint256 oldRate, uint256 newRate);
    
    /**
     * @dev Constructor - dipanggil saat deploy contract
     * @param initialSupply Jumlah token awal yang di-mint ke deployer (untuk liquidity)
     */
    constructor(uint256 initialSupply) ERC20("Telkom Token", "TELKOM") {
        // Mint initial supply untuk liquidity pool (bisa di-burn nanti)
        _mint(msg.sender, initialSupply * 10**decimals());
        emit TokensMinted(msg.sender, initialSupply * 10**decimals(), bytes32(0));
    }
    
    /**
     * @dev Return jumlah decimals (18)
     */
    function decimals() public view virtual override returns (uint8) {
        return _decimals;
    }

    /**
     * @dev Request top-up dengan bukti pembayaran IDR
     * @param amount Jumlah token yang ingin di-mint (dalam wei)
     * @param paymentProof IPFS hash atau URL bukti transfer bank
     */
    function requestTopup(uint256 amount, string memory paymentProof) public returns (bytes32) {
        require(amount >= minMintAmount, "Amount below minimum");
        require(bytes(paymentProof).length > 0, "Payment proof required");

        // Generate unique request ID
        bytes32 requestId = keccak256(abi.encodePacked(
            msg.sender,
            amount,
            paymentProof,
            block.timestamp,
            userNonces[msg.sender]++
        ));

        // Store request
        topupRequests[requestId] = TopupRequest({
            user: msg.sender,
            amount: amount,
            paymentProof: paymentProof,
            timestamp: block.timestamp,
            processed: false
        });

        emit TopupRequested(msg.sender, amount, requestId, paymentProof);
        return requestId;
    }

    /**
     * @dev Process top-up request (dapat dipanggil siapa saja untuk verifikasi)
     * @param requestId ID dari request yang akan diproses
     */
    function processTopup(bytes32 requestId) public {
        TopupRequest storage request = topupRequests[requestId];
        require(request.user != address(0), "Request not found");
        require(!request.processed, "Request already processed");
        require(block.timestamp >= request.timestamp + 1 hours, "Wait 1 hour before processing");

        // Mark as processed
        request.processed = true;

        // Mint tokens
        _mint(request.user, request.amount);
        emit TokensMinted(request.user, request.amount, requestId);
    }

    /**
     * @dev Instant topup for TLC Wallet (no waiting period)
     * @param amount Jumlah token yang akan di-mint
     * @param paymentProof Bukti pembayaran (untuk tracking)
     */
    function instantTopup(uint256 amount, string memory paymentProof) public returns (bytes32) {
        require(amount >= minMintAmount, "Amount below minimum");

        // Generate unique request ID
        bytes32 requestId = keccak256(abi.encodePacked(msg.sender, amount, block.timestamp, userNonces[msg.sender]++));

        // Store request for tracking
        topupRequests[requestId] = TopupRequest({
            user: msg.sender,
            amount: amount,
            paymentProof: paymentProof,
            timestamp: block.timestamp,
            processed: true // Already processed
        });

        // Mint tokens immediately
        _mint(msg.sender, amount);
        emit TokensMinted(msg.sender, amount, requestId);
        emit TopupRequested(msg.sender, amount, requestId, paymentProof);

        return requestId;
    }

    /**
     * @dev Request withdraw ke rekening bank
     * @param amount Jumlah token yang akan di-burn
     * @param bankAccount Nomor rekening tujuan (format: BANK_NAME:ACCOUNT_NUMBER:ACCOUNT_NAME)
     */
    function requestWithdraw(uint256 amount, string memory bankAccount) public returns (bytes32) {
        require(amount >= minBurnAmount, "Amount below minimum");
        require(balanceOf(msg.sender) >= amount, "Insufficient balance");
        require(bytes(bankAccount).length > 0, "Bank account required");

        // Generate unique request ID
        bytes32 requestId = keccak256(abi.encodePacked(
            msg.sender,
            amount,
            bankAccount,
            block.timestamp,
            userNonces[msg.sender]++
        ));

        // Burn tokens immediately (irreversible)
        _burn(msg.sender, amount);

        // Store withdraw request for processing
        withdrawRequests[requestId] = WithdrawRequest({
            user: msg.sender,
            amount: amount,
            bankAccount: bankAccount,
            timestamp: block.timestamp,
            processed: false
        });

        emit TokensBurned(msg.sender, amount, requestId);
        emit WithdrawRequested(msg.sender, amount, requestId, bankAccount);
        return requestId;
    }

    /**
     * @dev Transfer token dengan catatan (untuk tracking payment)
     * @param to Address tujuan
     * @param amount Jumlah token
     * @param note Catatan transaksi
     */
    function paymentTransfer(address to, uint256 amount, string memory note) public returns (bool) {
        _transfer(msg.sender, to, amount);
        emit PaymentProcessed(msg.sender, to, amount, note);
        return true;
    }

    /**
     * @dev Update exchange rate (governance function - bisa dipanggil siapa saja)
     * @param newRate Rate baru dalam wei (1e18 = 1 IDR)
     */
    function updateExchangeRate(uint256 newRate) public {
        require(newRate > 0, "Rate must be positive");
        uint256 oldRate = exchangeRate;
        exchangeRate = newRate;
        emit ExchangeRateUpdated(oldRate, newRate);
    }

    /**
     * @dev Get balance dengan format yang mudah dibaca
     * @param account Address yang ingin dicek
     */
    function getBalance(address account) public view returns (uint256) {
        return balanceOf(account);
    }

    /**
     * @dev Get topup request details
     * @param requestId ID dari request
     */
    function getTopupRequest(bytes32 requestId) public view returns (
        address user,
        uint256 amount,
        string memory paymentProof,
        uint256 timestamp,
        bool processed
    ) {
        TopupRequest memory request = topupRequests[requestId];
        return (request.user, request.amount, request.paymentProof, request.timestamp, request.processed);
    }

    /**
     * @dev Get withdraw request details
     * @param requestId ID dari request
     */
    function getWithdrawRequest(bytes32 requestId) public view returns (
        address user,
        uint256 amount,
        string memory bankAccount,
        uint256 timestamp,
        bool processed
    ) {
        WithdrawRequest memory request = withdrawRequests[requestId];
        return (request.user, request.amount, request.bankAccount, request.timestamp, request.processed);
    }

    /**
     * @dev Emergency burn function (anyone can burn their own tokens)
     * @param amount Jumlah token yang akan di-burn
     */
    function burn(uint256 amount) public {
        _burn(msg.sender, amount);
        emit TokensBurned(msg.sender, amount, bytes32(0));
    }
}