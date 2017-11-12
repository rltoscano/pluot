class PluotSplit extends PluotFormat(Polymer.Element) {
  static get is() { return 'pluot-split'; }
  static get properties() {
    return {
      splitTxn: { type: Object, value: { userDisplayName: "", amount: 0 } },
      splits: { type: Array, notify: true, value: [] },
      _remaining: { type: Number, computed: '_computeRemaining(splitTxn, splits.*)' },
    };
  }
  _computeRemaining(splitTxn, splitsProp) {
    return splitsProp.base.reduce(
        (remaining, split) => remaining - split.amount, splitTxn.amount);
  }
  _computeSplitButtonDisabled(_remaining, splitsProp) {
    return (_remaining != 0) ||
        splitsProp.base.some(s => !s.displayName || !s.category || !s.amount);
  }
  _onAddSplitTap(evt) {
    this.push('splits', { displayName: '', category: 1, amount: this._remaining });
  }
  _onRemoveSplitTap(evt) { this.splice('splits', evt.currentTarget.dataset.index, 1); }
  _dispatchDone() { this.dispatchEvent(new CustomEvent('done')); }
}

window.customElements.define(PluotSplit.is, PluotSplit);
