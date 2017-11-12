class PluotTxn extends PluotFormat(Polymer.Element) {
  static get is() { return 'pluot-txn'; }
  static get properties() {
    return {
      txn: { type: Object, value: null, observer: '_updateModel' },
      _selectingCategory: { type: Boolean, value: false, observer: '_dispatchResize' },
      _editingName: { type: Boolean, value: false },
      _category: { type: Number, value: 1, observer: '_maybeDispatchCategory' },
      _name: { type: String, value: null },
    };
  }
  _updateModel(txn) {
    if (txn != null) {
      this.setProperties({
        _category: txn.userCategory || txn.category,
        _name: this._formatDisplayName(txn),
      });
    }
  }
  _dispatchNameChanged(evt) {
    console.log('Name changed: ' + evt.target.value);
    this.dispatchEvent(
        new CustomEvent(
            'edit', {detail: {id: this.txn.id, field: 'name', value: evt.target.value }}));
  }
  _dispatchSplit() {
    console.log('Split requested: ' + this.txn.id);
    this.dispatchEvent(new CustomEvent('split', {detail: { id: this.txn.id }}));
  }
  _dispatchResize() { this.dispatchEvent(new CustomEvent('resize')); }
  _maybeDispatchCategory(newValue) {
    if (this.txn && newValue !== (this.txn.userCategory || this.txn.category)) {
      console.log('Category selection changed: ' + newValue);
      this.dispatchEvent(
          new CustomEvent(
              'edit', {detail: {id: this.txn.id, field: 'category', value: newValue }}));
    }
  }
}

window.customElements.define(PluotTxn.is, PluotTxn);
