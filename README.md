# Pectra CLI Validators Tool

A command-line interface (CLI) to streamline Ethereum validator operations such as batch switching of withdrawal credentials, consolidating validators, and performing execution layer (EL) exits (both partial and full).

## ✨ Features

*   **Batch Switch**: Update deposit credentials for multiple validators in a single transaction.
*   **Batch Consolidate**: Consolidate funds from multiple source validators to a single target validator.
*   **Batch EL Exit**: Perform partial or full exits for multiple validators from the execution layer.
*   **Secure**: Prompts for private key input securely at runtime; it's not stored in configuration files.
*   **Dynamic Fee Calculation**: Automatically fetches required fees per validator from the smart contract.

## 🚀 Installation

1.  Ensure you have Go (version 1.21+ recommended, as per `.github/workflows/release.yml` line 18) installed on your system.
2.  Clone the repository (if you haven't already).
3.  Build the CLI tool:

    ```bash
    go build -o pectra-cli ./cmd/main.go
    ```

    This will create an executable file named `pectra-cli` in the current directory.

## ⚙️ Configuration

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

*   `rpcUrl` (string): The URL of your Ethereum execution client RPC endpoint.
*   `blockExplorerUrl` (string): The base URL for your preferred block explorer (e.g., `https://etherscan.io`). Used for displaying transaction links.
*   `pectraBatchContract` (string): The address of the deployed Pectra batch contract.
*   `switch.validators` (array of strings): A list of validator public keys (hexadecimal, no "0x" prefix) for the batch switch operation.
*   `consolidate.sourceValidators` (array of strings): A list of source validator public keys for the batch consolidation operation.
*   `consolidate.targetValidator` (string): The target validator public key for consolidation.
*   `elExit.validators` (object): A map where keys are validator public keys and values are objects containing:
    *   `amount` (number): The amount in **Gwei** to withdraw for a partial exit. For a full exit, set this to `0`.
    *   `confirmFullExit` (boolean): Must be `true` if `amount` is `0` to confirm a full exit. Otherwise, `false`.

## 🔑 Private Key Handling

For security, your withdrawal address private key is **not** stored in the `config.json` file. The CLI will securely prompt you to enter it at runtime when an operation is initiated.
(See `internal/config/config.go` lines 88-114)

## 📜 ABI Dependency

The CLI requires the Pectra batch contract's ABI. Place the `abi.json` file in the same directory as the `pectra-cli` executable.
(The ABI is loaded from `./abi.json` as seen in `cmd/main.go` line 49)

## Unset Delegation

> It is HIGHLY recommended to unset delegation after performing any operation. This helps with restoration of EOA functionality and prevents the address from being used as a smart contract.

To unset delegation for a validator, run:
```bash
./pectra-cli unset-delegation config.json
```


## 🛠️ Usage

The general command structure is:
```bash
./pectra-cli <command> config.json
```

To get a gist of the CLI, run:
```bash
./pectra-cli --help
```

Replace `<command>` with one of the operations listed below and `config.json` with the path to your configuration file.

### Switch Validators

Updates deposit credentials for the validators specified in `config.json` under the `switch` section.

```bash
./pectra-cli switch config.json
```



### Consolidate Validators

Consolidates funds from `sourceValidators` to `targetValidator` as specified in `config.json` under the `consolidate` section.

```bash
./pectra-cli consolidate config.json
```


### Execution Layer (EL) Exit

Performs partial or full exits for validators specified in `config.json` under the `elExit` section.

```bash
./pectra-cli el-exit config.json
```


## 📝 Important Notes

*   **Validator Public Keys**: All validator public keys in the `config.json` file must be in hexadecimal format, without the "0x" prefix.
*   **Transaction Fees**: The fee required per validator for each operation (switch, consolidate, EL exit) is automatically fetched from the smart contract functions (`getConsolidationFee`, `getExitFee`). This fee is in Wei. The total transaction `value` sent will be `(number of validators) * (fee per validator)`.
    (Fee fetching logic: `cmd/main.go` lines 67-74, 78-79, 91-92, 105-106) and `internal/utils/utils.go` lines 98-125
*   **Execution Layer (EL) Exits**:
    *   The `amount` specified in the `elExit.validators` section of `config.json` is in **Gwei** (1 ETH = 1,000,000,000 Gwei).
        (See `internal/utils/utils.go` for usage notes, and `internal/operations/partialexit.go` lines 41-49 for handling)
    *   For a **full exit**, set `amount` to `0` (or `0.0`) and `confirmFullExit` to `true`.
    *   For a **partial exit**, specify the desired `amount` in Gwei (e.g., `10.0` for 10 Gwei) and ensure `confirmFullExit` is `false`.
*   **Transaction Authorization**: This tool utilizes EIP-7702 SetCode transaction authorization for its operations.
    (See `internal/transaction/transaction.go` lines 18-61)

## 🖼️ Sample Output

(The CLI provides informative output during its execution, including connection status, fees, transaction hashes, and success/failure messages.)

![Sample Output](https://i.imgur.com/niLux60.png)
*(Note: The sample output image might show older field names or values; refer to the current configuration guidelines.)*

## 🏗️ Project Structure

The project is organized as follows:

*   `cmd/main.go`: The entry point for the CLI application.
*   `internal/`: Contains the core logic of the application.
    *   `config/`: Handles loading and validation of the `config.json` file and ABI.
    *   `operations/`: Implements the logic for switch, consolidate, and EL exit operations.
    *   `transaction/`: Manages the creation, signing, and sending of Ethereum transactions.
    *   `utils/`: Provides utility functions, including printing usage information and fee fetching.
*   `abi.json`: The contract ABI (must be present at runtime).
*   `.github/workflows/`: Contains GitHub Actions workflows for CI/CD, including automated releases.

## 🔄 Building and Releasing

The project uses GitHub Actions for automated builds and releases. When code is pushed to the `main` branch:
1.  Linters are run.
2.  Binaries are built for Linux, macOS, and Windows.
3.  A new GitHub release is created with the version (date-based + short commit SHA) and the compiled binaries are attached as assets.
    (See `.github/workflows/release.yml`)

---

Feel free to suggest improvements or report issues!
