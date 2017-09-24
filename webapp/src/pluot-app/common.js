function formatAmount(amount) {
  if (amount == 0) {
    return "$0.00";
  }
  var amountStr = String(amount);
  return "$" + amountStr.substr(0, amountStr.length-2) + "." + amountStr.substr(-2);
}

function formatDate(str) {
  var d = new Date(Date.parse(str));
  return d.toLocaleString('en-US', { month: 'short', day: 'numeric' });
}
