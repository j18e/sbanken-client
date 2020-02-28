function addToSelectors() {
  sel = document.getElementById('date-selector');
  rows = document.getElementById('spending-table').getElementsByClassName("table-row");
  dates = [];

  for (var i = 0; i < rows.length; i++) {
    date = rows[i].getElementsByClassName("date-cell")[0].textContent;
    if (dates.includes(date)) {
      continue
    }
    dates.push(date);
  }

  for (var i = 0; i < dates.length; i++) {
    opt = document.createElement('option');
    opt.value = dates[i];
    opt.innerHTML = dates[i];
    sel.appendChild(opt);
  }
}

function logSomething() {
  console.log('test log');
}

function selectOpt(sel) {
  val = sel.options[sel.selectedIndex].value;
  console.log(sel.options.length);
  rows = document.getElementById("spending-table").getElementsByClassName("table-row");
  for (var i = 0; i < rows.length; i++) {
    if (val == 'all') {
      rows[i].style.display = '';
    } else if (rows[i].getElementsByClassName("date-cell")[0].textContent != val) {
      rows[i].style.display = "none";
    } else {
      rows[i].style.display = '';
    }
  }
  // div.style.display = "none";
}
