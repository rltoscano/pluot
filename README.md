# Prerequisites

1. [Polymer 2.x](https://www.polymer-project.org/2.0/start/install-2-0)
2. [Google App Engine Standard SDK for Golang](https://cloud.google.com/appengine/docs/standard/go/download)

# Installation

1. Pull down all the polymer dependencies with `polymer install` run from the
   `webapp` folder.

# Deployment

To deploy to App Engine, you'll need to [create a Google Cloud
project](https://cloud.google.com/resource-manager/docs/creating-managing-projects).

1. Run `go run package_webapp.go` from the `tool` folder to build the web app and copy it to the
   service folder.
1. Run `dev_appserver.py service/src/github.com/rltoscano/pluot/app.yaml` to run
   a development server. The server will be started at http://localhost:8080
1. To deploy to prod run
   `gcloud app deploy service/src/github.com/rltoscano/pluot/app.yaml --project=$MYPROJECT`
   where $MYPROJECT is set to whatever App Engine app you've created.

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
- `Splits` - `int[]`, IDs of transactions split from this one.
- `LastUpdated` - `Timestamp`
- `SplitSourceID` - `int`, ID of the source transaction that this transaction was split from.

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

# UI

- Color scheme: https://www.computerhope.com/cgi-bin/htmlcolor.pl?c=583759
