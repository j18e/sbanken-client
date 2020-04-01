function addToSelectors(id, cellType) {
  sel = document.getElementById(id).querySelector('select');
  rows = document.querySelector('#spending-table').querySelector('tbody').querySelectorAll('tr');
  items = [];

  for (var i = 0; i < rows.length; i++) {
    item = rows[i].getElementsByClassName(cellType)[0].textContent;
    if (items.includes(item)) {
      continue
    }
    items.push(item);
  }

  for (var i = 0; i < items.length; i++) {
    opt = document.createElement('option');
    opt.value = items[i];
    opt.innerHTML = items[i];
    sel.appendChild(opt);
  }
}

function selectOpt(sel, cellType) {
  val = sel.options[sel.selectedIndex].value;
  rows = document.querySelector('#spending-table').querySelector('tbody').querySelectorAll('tr');

  for (var i = 0; i < rows.length; i++) {
    if (val == 'all') {
      rows[i].style.display = '';
    } else if (rows[i].getElementsByClassName(cellType)[0].textContent != val) {
      rows[i].style.display = "none";
    } else {
      rows[i].style.display = '';
    }
  }
  updateTotal();
}

function updateTotal() {
  var total = 0;
  // document.querySelector('.message').remove();
  sel = document.querySelector('#spending-total');
  rows = document.querySelector('#spending-table').querySelector('tbody').querySelectorAll('tr');
  for (var i = 0; i < rows.length; i++) {
    if (rows[i].style.display == 'none') {
      continue;
    }
    total += parseInt(rows[i].getElementsByClassName("nok-cell")[0].textContent);
  }
  sel.textContent = `Total: ${total} NOK`;
}
