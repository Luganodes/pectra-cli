# Pectra CLI

A command-line tool for Ethereum validator operations including batch consolidation, switching, and partial exits.

## Installation

```bash
go build -o pectra-cli
```

## Configuration

Create a JSON configuration file (`input.json`) with the following structure:

```json
{
    "withdrawalAddressPrivateKey": "",
    "rpcUrl": "http://100.103.129.92:8545",
    "pectraBatchContract": "0x1a91dDF19470354708cD302A6AA8892D50b20Ff5",
    "switch": {
        "validators": ["a2952c28b1a1ecbad6b9a6ad7452f607cd5c1210734d4dd1f8ce69cdb229d3da6b63244b0a960b9d7818814dfbf29bc8"],
        "amountPerValidator": 1
    },
    "consolidate": {
        "sourceValidators": ["a2952c28b1a1ecbad6b9a6ad7452f607cd5c1210734d4dd1f8ce69cdb229d3da6b63244b0a960b9d7818814dfbf29bc8"],
        "targetValidator": "a040ee785f2d78d3ae6ca41c20327b6b93d531302db98179fc127438ec2f96b1cc2993198ea2aff4c4831f0a2c243b57",
        "amountPerValidator": 1
    },
    "partialExit": {
        "validators": {
            "a2952c28b1a1ecbad6b9a6ad7452f607cd5c1210734d4dd1f8ce69cdb229d3da6b63244b0a960b9d7818814dfbf29bc8": 2000000000000000000,
            "abf19e3615da01ca9b6cc7915fc5074b88ee80f63f1ef5977649082614f89b65c0969918ca30465212c34ca217dae9a7": 3000000000000000000
        },
        "amountPerValidator": 1
    }
}
```

> Note: All the validator addresses mentioned here are just for example. You can use your own validator addresses.

## Dependencies

The CLI requires `abi.json` in the same directory, which contains the ABI for the Pectra batch contract.

## Usage

### Switch Validators

```bash
./pectra-cli switch input.json
```

### Consolidate Validators

```bash
./pectra-cli consolidate input.json
```

### Partial Exit

```bash
./pectra-cli partial-exit input.json
```

### Sample Output

![Ouput](https://i.imgur.com/niLux60.png)

## Notes

- The `withdrawalAddressPrivateKey` should be a hexadecimal string without the "0x" prefix
- All validator pubkeys should be in hexadecimal format without the "0x" prefix
- `amountPerValidator` values are in Wei (10^18)
- The CLI will automatically convert Wei amounts to Gwei for partial exits
- You can customize the `amountPerValidator` field for each operation:
  - For `switch`: Value sent with each validator operation (defaults to 1 if not specified)
  - For `consolidate`: Value sent with the consolidation operation (defaults to 1 if not specified)
  - For `partialExit`: Value sent with the partial exit operation (defaults to 1 if not specified)
