# Pectra CLI Validators Tool

![banner](https://i.imgur.com/7YKiU9D.png)

A powerful airgapped CLI tool for executing Ethereum validator operations including consolidation, switching, and both partial and full withdrawals with seamless batching enabled by EIP-7702.

# Background

The Pectra upgrade (Prague + Electra) introduces key validator enhancements: consolidation allows merging multiple validators into one to simplify management and reduce overhead; switch enables validators to update their BLS keys, supporting strategies like autocompounding without needing to exit and re-enter; and execution layer exits allow validators to exit directly via the execution layer, streamlining the exit process and enabling better integration with smart contracts and tooling.

This CLI leverages the Pectra.sol smart contract, designed to facilitate batch operations for Ethereum validator management tasks resulting in 60–80% gas savings. It leverages EIP-7702 (Set EOA account code) to allow validator withdrawal Externally Owned Accounts (EOAs) to execute multiple operations (consolidation, credential switching, Execution Layer exits) in a single, atomic transaction.

This overcomes the limitation of the official Ethereum Foundation "system assembly" (sys-asm) contracts, which only permit one operation per transaction and require direct initiation by the withdrawal EOA.

Moreover, the CLI is designed to be airgapped, ensuring security by allowing users to generate unsigned transactions without exposing sensitive keys. These transactions can then be signed externally using the appropriate withdrawal address in a secure, offline environment, and broadcasted manually to the network. This approach supports safe validator operations without compromising key custody.

## Features

- **Batch Switch**: Update withdrawal credentials for multiple validators from 0x01 type to 0x02 type in a single transaction.
- **Batch Consolidate**: Consolidate funds from multiple source validators to a single target validator in a single transaction.
- **Batch EL Exit**: Perform partial or full exits for multiple validators from the execution layer.
- **Secure**: Supports airgapped workflows by generating unsigned transactions that users can sign with their withdrawal address in a secure offline environment.
- **Dynamic Fee Calculation**: Automatically fetches required fees per validator from the smart contract.
- **Secure, audited contract**: The Pectra batch contract is open-source and has been independently audited by [Quantstamp](https://quantstamp.com). <br>
  🔗 [View the contract repository](https://github.com/Luganodes/Pectra-Batch-Contract) <br>
  🛡️ [Read the full audit report](https://github.com/Luganodes/Pectra-Batch-Contract/blob/main/audits/quantstamp/Audit.pdf) <br>
  📄 [Audit Certificate](https://certificate.quantstamp.com/full/luganodes-pectra-batch-contract/23f0765f-969a-4798-9edd-188d276c4a2b/index.html)

## Installation

1.  Ensure you have Go (version 1.21+ recommended, as per `.github/workflows/release.yml` line 18) installed on your system.
2.  Clone the repository (if you haven't already).
3.  Build the CLI tool:

    ```bash
    go build -o pectra-cli ./cmd/main.go
    ```

    This will create an executable file named `pectra-cli` in the current directory.

## Deployed Contracts
- Mainnet - [0x17c11FDdADac2b341F2455aFe988fec4c3ba26e3](https://etherscan.io/address/0x17c11FDdADac2b341F2455aFe988fec4c3ba26e3)
- Hoodi - [0xe264B0F3e491Ab5aEd2C0A32956cb9e68707F457](https://hoodi.etherscan.io/address/0xe264B0F3e491Ab5aEd2C0A32956cb9e68707F457)

## Configuration

Create a JSON configuration file named `config.json` in the same directory as the `pectra-cli` executable. You can use `sample_config.json` as a template.

**Example `config.json`:**

```json
{
  "rpcUrl": "http://100.71.214.23:8545",
  "blockExplorerUrl": "https://hoodi.etherscan.io",
  "pectraBatchContract": "0x209eF6e6d26953E30B652300Ac4a0A5De90f79F6",
  "switch": {
    "validators": [
      "b5a2635ef8d420a0c5d23341c638dd11a500aefa8f7d9fc1f726edbf8163f4e0b727f47faa57b91af50c13e863f13142",
      "b5f27dae0d6623d953252405056ce3a56ddf575de95c46cb212d396ffe4ccf0f905c138a897e8bde2e7e146705f88306",
      "880165dbbc70136744d942a450317d9e2cb4684eb460e33b0171ccfa4ddac99eb93c5603b95511d0f7b388f47ebcd36f"
    ]
  },
  "consolidate": {
    "sourceValidators": [
      "880165dbbc70136744d942a450317d9e2cb4684eb460e33b0171ccfa4ddac99eb93c5603b95511d0f7b388f47ebcd36f",
      "b5a2635ef8d420a0c5d23341c638dd11a500aefa8f7d9fc1f726edbf8163f4e0b727f47faa57b91af50c13e863f13142"
    ],
    "targetValidator": "b5f27dae0d6623d953252405056ce3a56ddf575de95c46cb212d396ffe4ccf0f905c138a897e8bde2e7e146705f88306"
  },
  "elExit": {
    "validators": {
      "b5f27dae0d6623d953252405056ce3a56ddf575de95c46cb212d396ffe4ccf0f905c138a897e8bde2e7e146705f88306": {
        "amount": 1000000000,
        "confirmFullExit": false
      },
      "a64d428e5933b2a54c64431ab9f99dcd1bc943cb6c38de9aadb221a2af6167ab9caccc384ae221d738886942daf15788": {
        "amount": 0,
        "confirmFullExit": true
      }
    }
  }
}
```

**Configuration Fields:**

- `rpcUrl` (string): The URL of your Ethereum execution client RPC endpoint.
- `blockExplorerUrl` (string): The base URL for your preferred block explorer (e.g., `https://etherscan.io`). Used for displaying transaction links.
- `pectraBatchContract` (string): The address of the deployed Pectra batch contract.
- `switch.validators` (array of strings): A list of validator public keys (hexadecimal, no "0x" prefix) for the batch switch operation. Maximum source validators for switch: 200
- `consolidate.sourceValidators` (array of strings): A list of source validator public keys with 0x01 type withdrawal credentials for the batch consolidation operation. Maximum validators for consolidation: 63
- `consolidate.targetValidator` (string): The target validator public key for consolidation. Consolidated stake must be less than or equal to 2048 ETH otherwise surplus stake will get automatically sweeped.
- `elExit.validators` (object): A map where keys are validator public keys ( maximum of 200 ) and values are objects containing:
  - `amount` (number): The amount in **Gwei** to withdraw for a partial exit. For a full exit, set this to `0`.
  - `confirmFullExit` (boolean): Must be `true` if `amount` is `0` to confirm a full exit. Otherwise, `false` for partial exit and such an `amount` where remaining balance after the exit is at least 32 ETH.

⚠️ Ensure only required validator addresses are set in config.json and their corresponding private keys are provided via the CLI — missing or incorrect entries may result in unintended transfer of funds. <br><br>

## Private Key Handling

Use the `--airgapped` or `-a` to run the CLI in airgapped mode; alternatively, omit the flag to sign directly in the CLI by providing the private key. The CLI will securely prompt you to enter it at runtime when an operation is initiated.
(See `internal/config/config.go` lines 88-114)

⚠️ Ensure that correct private keys are provided for the validators — otherwise, transactions will succeed but no validator operation will occur, wasting gas. <br><br>

## 📜 ABI Dependency

The CLI requires the Pectra batch contract's ABI. Place the `abi.json` file in the same directory as the `pectra-cli` executable.
(The ABI is loaded from `./abi.json` as seen in `cmd/main.go` line 49) <br><br>

## ⚠️ Unset Delegation

> It is HIGHLY recommended to unset delegation after performing any operation. This helps with restoration of EOA functionality and prevents the address from being used as a smart contract.

To unset delegation for a validator, run:

```bash
./pectra-cli unset-delegation -c config.json
```

## Usage

The general command structure is:

```bash
./pectra-cli <command> -c config.json 
```

To get a gist of the CLI, run:

```bash
./pectra-cli --help
```

Replace `<command>` with one of the operations listed below and `config.json` with the path to your configuration file.

Add the `-a` or `--airgapped` flag to run the CLI in airgapped mode.

### Switch Validators

Updates deposit credentials for the validators specified in `config.json` under the `switch` section. You can switch up to 200 validators in a single batch.

```bash
./pectra-cli switch -c config.json
```

⚠️ Do not switch a validator that has already been switched — the transaction will succeed but the switch won't take effect, wasting gas. <br><br>

### Consolidate Validators

Consolidates funds from `sourceValidators` to `targetValidator` as specified in `config.json` under the `consolidate` section. You can consolidate from up to 63 source validators into one target validator.

```bash
./pectra-cli consolidate -c config.json
```

⚠️ Do not use exited validators as source or target — transactions will succeed but consolidation won't occur, wasting gas.<br><br>

### Execution Layer (EL) Exit

Performs partial or full exits for validators specified in `config.json` under the `elExit` section. You can exit up to 200 validators in a single batch.

```bash
./pectra-cli el-exit -c config.json
```

⚠️ Do not attempt to exit a validator that has already exited — the transaction will succeed but no exit will occur, wasting gas. <br><br>

### Signing and Broadcast for airgapped mode

To sign an unsigned transaction, use `scripts/sign.go` on the `unsigned_txn.json` file — this will generate a `signed_txn.json`. 

```bash
go run scripts/sign.go unsigned_txn.json
```
Once signed, use the CLI's broadcast command to submit the `signed_txn.json` to the network.


```bash
./pectra-cli broadcast -c config.json -f signed_txn.json
```

## 📝 Important Notes

- **Validator Public Keys**: All validator public keys in the `config.json` file must be in hexadecimal format, without the "0x" prefix.
- **Transaction Fees**: The fee required per validator for each operation (switch, consolidate, EL exit) is automatically fetched from the smart contract functions (`getConsolidationFee`, `getExitFee`). This fee is in Wei. The total transaction `value` sent will be `(number of validators) * (fee per validator)`.
  (Fee fetching logic: `cmd/main.go` lines 67-74, 78-79, 91-92, 105-106) and `internal/utils/utils.go` lines 98-125
- **Execution Layer (EL) Exits**:
  - The `amount` specified in the `elExit.validators` section of `config.json` is in **Gwei** (1 ETH = 1,000,000,000 Gwei).
    (See `internal/utils/utils.go` for usage notes, and `internal/operations/partialexit.go` lines 41-49 for handling)
  - For a **full exit**, set `amount` to `0` (or `0.0`) and `confirmFullExit` to `true`.
  - For a **partial exit**, specify the desired `amount` in Gwei (e.g., `10.0` for 10 Gwei) and ensure `confirmFullExit` is `false`.
- **Transaction Authorization**: This tool utilizes EIP-7702 SetCode transaction authorization for its operations.
  (See `internal/transaction/transaction.go` lines 18-61)

<br>

## Sample Output

(The CLI provides informative output during its execution, including connection status, fees, transaction hashes, and success/failure messages.)
### Sample Output for airgapped mode
![Sample Airgapped Output](https://i.imgur.com/CfKpNsN.png)
<br>
![Sample Signing Output](https://i.imgur.com/M9GGQB0.png)
<br>
![Sample Broadcast Output](https://i.imgur.com/nIEZDl8.png)
<br>
### Sample Output for non-airgapped mode
![Sample Non Airgapped Output](https://i.imgur.com/L2wpleY.png)
_(Note: The sample output images might show older field names or values; refer to the current configuration guidelines.)_

<br>

## Project Structure

The project is organized as follows:

- `cmd/main.go`: The entry point for the CLI application.
- `internal/`: Contains the core logic of the application.

  - `config/`: Handles loading and validation of the `config.json` file and ABI.
  - `operations/`: Implements the logic for switch, consolidate, and EL exit operations.
  - `transaction/`: Manages the creation, signing, and sending of Ethereum transactions.
  - `utils/`: Provides utility functions, including printing usage information and fee fetching.

- `.github/workflows/`: Contains GitHub Actions workflows for CI/CD, including automated releases.

<br>

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feat/AmazingFeature`)
3. Make some amazing changes.
4. `git add .`
5. Commit your Changes (`git commit -m "<Verb>: <Action>"`)
6. Push to the Branch (`git push origin feat/AmazingFeature`)
7. Open a Pull Request

To start contributing, check out [`CONTRIBUTING.md`](./CONTRIBUTING.md) . New contributors are always welcome to support this project.

## License

Distributed under the MIT License. See [`LICENSE`](./LICENSE) for more information.
