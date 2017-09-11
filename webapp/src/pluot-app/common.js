function formatAmount(amount) {
  var amountStr = String(amount);
  return "$" + amountStr.substr(0, amountStr.length-2) + "." + amountStr.substr(-2);
}
