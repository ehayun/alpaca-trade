// All index symbols (S&P 500, Nasdaq-100, Dow 30) with a search box and a market buy on each row.
// The list loads the first time the tab opens.
import { api, num, showMessage, hideMessage, cell, pickable, markShown } from './common.js';

const $ = (id) => document.getElementById(id);
const body = $('all-body');
const msg = $('all-msg');
const search = $('all-search');
const DEFAULT_QTY = 10; // same as the scanner
let shown = ''; // symbol on the chart
let loaded = false;

async function load() {
  loaded = true;
  hideMessage(msg);
  try {
    const { symbols } = await api('/api/symbols');
    body.replaceChildren(...symbols.map(row));
    markShown(body, shown);
    filter();
  } catch (err) {
    loaded = false; // try again on the next tab click
    showMessage(msg, err.message);
  }
}

function row(symbol) {
  const tr = document.createElement('tr');
  pickable(tr, symbol);
  const qty = document.createElement('input');
  qty.type = 'number';
  qty.className = 'form-control form-control-sm scan-qty';
  qty.dir = 'ltr';
  qty.min = '0';
  qty.step = 'any';
  qty.value = DEFAULT_QTY;
  qty.setAttribute('aria-label', `כמות ${symbol}`);
  qty.addEventListener('click', (e) => e.stopPropagation()); // edit only, do not open the symbol

  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'btn btn-sm btn-outline-success py-0';
  btn.textContent = 'קנה';
  btn.setAttribute('aria-label', `קנה ${symbol} במחיר שוק`);
  btn.addEventListener('click', (e) => {
    e.stopPropagation(); // buy only, do not open the symbol
    buy(symbol, qty.value, btn);
  });

  const btnCell = document.createElement('td');
  btnCell.append(btn);
  const qtyCell = document.createElement('td');
  qtyCell.append(qty);
  tr.append(btnCell, cell(symbol, { ltr: true }), qtyCell);
  return tr;
}

async function buy(symbol, qty, btn) {
  if (!confirm(`לקנות ${qty} ${symbol} במחיר שוק?`)) return;
  hideMessage(msg);
  btn.disabled = true;
  try {
    const o = await api('/api/orders', {
      method: 'POST',
      body: JSON.stringify({ symbol, side: 'buy', type: 'market', qty }),
    });
    showMessage(msg, `נשלחה קנייה במחיר שוק: ${num(o.qty)} ${o.symbol}`, 'success');
    document.dispatchEvent(new Event('orders-changed')); // trade.js reloads positions and orders
  } catch (err) {
    showMessage(msg, err.message);
  } finally {
    btn.disabled = false;
  }
}

// filter shows the rows whose symbol starts with the search text.
function filter() {
  const q = search.value.trim().toUpperCase();
  let n = 0;
  for (const tr of body.children) {
    tr.hidden = !tr.dataset.symbol.startsWith(q);
    if (!tr.hidden) n++;
  }
  $('all-empty').classList.toggle('d-none', n > 0 || !loaded);
}

search.addEventListener('input', filter);
$('tab-all').addEventListener('click', () => {
  if (!loaded) load();
});
document.addEventListener('symbol-loaded', (e) => {
  shown = e.detail;
  markShown(body, shown);
});
