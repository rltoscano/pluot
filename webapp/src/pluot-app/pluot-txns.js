class PluotTxns extends PluotFormat(Polymer.Element) {
  static get is() { return 'pluot-txns'; }
  static get properties() {
    return {
      mode: { type: String, notify: true, value: 'list' },
      query: { type: Object, value: null },
      _addDate: { type: String, value: new Date().toISOString().substring(0, 10) },
      _listResp: { type: Object, value: { txns: [] } },
      _splitTxn: { type: Object, value: { id: 0 } },
      _splits: { type: Array, value: [] },
      _txnModel: { type: Object, computed: '_computeTxnModel(query, _listResp.txns.*)' },
    };
  }
  _patch(evt) {
    this.$.list.notifyResize();
    console.log('Patching..');
    var body = { txn: {id: evt.detail.id}, fields: [] };
    if (evt.detail.field === 'name') {
      body.fields.push('userDisplayName');
      body.txn.userDisplayName = evt.detail.value;
    } else if (evt.detail.field === 'category') {
      body.fields.push('userCategory');
      body.txn.userCategory = evt.detail.value;
    } else {
      alert('Unexpected transaction edit');
      return;
    }
    this.$.patchAjax.url = this.$.globals.urlPrefix + 'txns/' + evt.detail.id;
    this.$.patchAjax.body = body;
    this.$.patchAjax.generateRequest();
  }
  _computeCreateReq(_addAmount, _addCategory, _addDate, _addDescription) {
    if (!_addAmount || !_addCategory || !_addDate || !_addDescription) {
      return {};
    }
    return {
      txn: {
        amount: _addAmount,
        userCategory: _addCategory,
        postDate: new Date(_addDate).toUTCString(),
        userDisplayName: _addDescription,
      }
    };
  }
  _computeTxnModel(query, txnsProp) {
    var txns = txnsProp.base;
    var model = { txnById: {}, filtered: [] };
    if (!txns) {
      return model;
    }
    txns.forEach(t => {
      model.txnById[t.id] = t;
      if (this._filter(t, query)) {
        model.filtered.push(t);
      }
    });
    return model;
  }
  _computeSplitReq(_splitTxn, _splitsProp) {
    return { sourceId: _splitTxn.id, splits: _splitsProp.base };
  }
  _createTxn(evt) {
    this.$.createAjax.generateRequest();
    this.mode = 'list';
  }
  _toastSplitSuccess(evt) {
    this.$.toast.text = 'Transaction split';
    this.$.toast.open();
    // TODO(robert): Apply changes to model.
  }
  _updateTxnListAndToast(evt) {
    for (var i = 0; i < this._listResp.txns.length; i++) {
      if (this._listResp.txns[i].id == evt.target.lastResponse.id) {
        this.splice('_listResp.txns', i, 1, evt.target.lastResponse);
        break;
      }
    }
    console.log('Txn patched: ' + evt.detail);
    this.$.toast.text = 'Transaction updated';
    this.$.toast.open();
  }
  _toastTxnEditFailed(evt) {
    this.$.toast.text = 'Transaction update failed';
    this.$.toast.open();
  }
  _sendSplitRequest(evt) {
    this.mode = 'list';
    this.$.splitAjax.generateRequest();
  }
  _showSplit(evt) {
    var splitSource = this._txnModel.txnById[evt.detail.id];
    if (splitSource.splitSourceId) {
      splitSource = this._txnModel.txnById[splitSource.splitSourceId];
    }
    this._splitTxn = splitSource;
    this._splits =
        (splitSource.splits || [])
            .map(id => this._txnModel.txnById[id])
            .map(s => (
                {
                  displayName: s.userDisplayName,
                  category: s.userCategory,
                  amount: s.amount
                }));
    this.mode = 'split';
  }
  _filter(txn, query) {
    if (txn.splits && txn.splits.length > 0) { // Hide split sources.
      return false;
    }
    if (query == null) {
      return true;
    }
    if (query.categories.length > 0) {
      if (!query.categories.includes(txn.userCategory || txn.category)) {
        return false;
      }
    }
    var postDate = new Date(txn.postDate);
    if (query.after != null && query.after.getTime() > postDate.getTime()) {
      return false;
    }
    if (query.before != null && query.before.getTime() <= postDate.getTime()) {
      return false;
    }
    if (query.isExpense != null &&
        query.isExpense == this.$.globals.isIncome(txn.userCategory || txn.category)) {
      return false;
    }
    return true;
  }
  _resizeItem(evt) {
    if (evt.target.txn) {
      this.$.list.updateSizeForItem(evt.target.txn);
    }
  }
  _toastAndInsertTransaction(evt) {
    this.$.toast.text = "Transaction created";
    this.$.toast.open();
    this.unshift('_listResp.txns', evt.target.lastResponse);
  }
  _toastCreateTxnFailed(evt) {
    this.$.toast.text = 'Transaction creation failed';
    this.$.toast.open();
  }
}

window.customElements.define(PluotTxns.is, PluotTxns);
