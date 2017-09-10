# Data Model

## Txn
Financial transactions.

Properties
- id - int64
- post_date - Timestamp
- amount - Currency
- original_display_name - string
- display_name - string
- user_display_name - string
- note - string
- category - Category
- user_category
- split - TxnSplit[]
- last_updated - Timestamp

## TxnSplit
A split of one transaction for the purpose of categorization.

Properties
- amount - Currency
- display_name - string
- note - string

## UploadEvent

Properties
- `id` - string
- `event_time` - Timestamp
- `user` - string
- `source` - string
- `start` - Timestamp
- `end` - Timestamp

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
- csv_data - Text of the CSV file.
- parser - How to interpret the data.

## ComputeAggregation
Calculates aggregations for a time-bound filtered list of transactions.

Parameters
- `start` - Timestamp
- `end` - Timestamp
- `category` - integer

Results
- `avg` - integer
- `total` - integer

# UI

- Color scheme: https://www.computerhope.com/cgi-bin/htmlcolor.pl?c=583759

# TODO

- [ ] Prevent CSRF attacks.
