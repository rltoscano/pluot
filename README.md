# Data Model

## Currency
All amounts are represented as 64-bit integers in US pennies and displayed in USD.

## Txn
Financial transactions.

Properties
- `ID` - `int64`
- `PostDate` - `Timestamp`
- `Amount` - `int64`
- `OriginalDisplayName` - `string`, name of the transaction as imported
- `DisplayName` - `string`, server-generated display friendly name of the transaction
- `UserDisplayName` - `string`
- `Note` - `string`
- `Category` - `int`, system-generated category.
- `UserCategory` - `int`, user override category.
- `Split` - `TxnSplit[]`
- `LastUpdated` - `Timestamp`

## TxnSplit
A split of one transaction for the purpose of categorization.

Properties
- `Amount` - `int64`
- `DisplayName` - `string`
- `Category` - `int`
- `Note` - `string`

## UploadEvent

Properties
- `ID` - `string`
- `EventTime` - `Timestamp`
- `User` - `string`
- `Source` - `string`
- `Start` - `Timestamp`
- `End` - `Timestamp`

# Processes

## Transaction Upload
1. Download CSV from bank.
1. Visit /upload.
1. Choose appropriate parser.
1. De-dupe any transactions.
1. Visit /txns to see all transactions.


# API

## Upload

Parameters
- `Csv` - `string`, text of the CSV file.
- `Source` - `string`, source of the uploaded transactions. E.g. bank.
- `Start` - `Timestamp`, starting bound of the upload which may before the earliest transaction in the upload
- `End` - `Timestamp`, ending bound of the upload which may after the most recent transaction in the upload

## ComputeAggregation
Calculates aggregations for a time-bound filtered list of transactions.

Parameters
- `start` - `Timestamp`
- `end` - `Timestamp`
- `category` - `int`

Results
- `Avg` - `int`
- `Total` - `int`

# UI

- Color scheme: https://www.computerhope.com/cgi-bin/htmlcolor.pl?c=583759

# TODO

- [ ] Prevent CSRF attacks.
- [ ] Create globals for host in webapp urls.
- [ ] Create globals for category lists.
