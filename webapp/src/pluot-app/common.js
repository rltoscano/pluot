function formatAmount(amount) {
  var amountStr = String(amount);
  return "$" + amountStr.substr(0, amountStr.length-2) + "." + amountStr.substr(-2);
}

function formatCategory(category) {
  switch (category) {
    case 1: return 'Uncategorized';
    case 2: return 'Entertainment';
    case 3: return 'Eating Out';
    case 4: return 'Groceries';
    case 5: return 'Shopping';
    case 6: return 'Health';
  }
}

function formatDate(str) {
  var d = new Date(Date.parse(str));
  return d.toLocaleString('en-US', { month: 'short', day: 'numeric' });
}
