/**
 * @customElement
 * @polymer
 */
class PluotApp extends Polymer.Element {
  static get is() { return 'pluot-app'; }
  static get properties() {
    return {
      txnsSearching: { type: Boolean, value: false, observer: '_clearTxnSearch' },
      subroute: { type: Object, observer: '_updateTxnSearch' },
      queryStr: { type: String, observer: '_updateSubrouteToQueryStr' },
    };
  }
  navToGitHubIssues() {
    window.location = "https://github.com/rltoscano/pluot/issues/new";
  }
  ready() {
    super.ready();
    if (!this.route.path) {
      this.set('route.path', "/agg");
    }
  }
  _eq(l, r) { return l == r; }
  _setTxnsModeToAdd(evt) { this.$.txnsList.mode = 'add'; }
  _setTxnsModeToList(evt) { this.$.txnsList.mode = 'list'; }
  _txnsModeNotListOrNotSearching(txnsMode, txnsSearching) {
    return txnsMode != 'list' || !txnsSearching;
  }
  _updateTxnSearch(subroute) {
    if (subroute.path && subroute.path.substr(1)) {
      this.txnsSearching = true;
      this.$.txnSearch.queryStr = subroute.path.substr(1);
    } else {
      this.txnsSearching = false;
    }
  }
  _updateSubrouteToQueryStr(queryStr) {
    var subroutePath;
    if (queryStr) {
      subroutePath = '/' + queryStr;
    } else {
      subroutePath = '';
    }
    this.set('subroute.path', subroutePath);
  }
  _closeDrawer() { this.$.drawer.opened = false; }
  _clearTxnSearch() { this.$.txnSearch.queryStr = ''; }
  _rulesEditNew() { this.$.rules.editNew(); }
  _rulesDelete() { this.$.rules.delete(); }
  _rulesConfirmEdit() { this.$.rules.confirmEdit(); }

}

window.customElements.define(PluotApp.is, PluotApp);
