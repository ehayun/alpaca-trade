// Watchlist: the user's own symbols, kept in an Alpaca watchlist. The list loads the first time the tab opens.
import { api, showMessage, hideMessage, cell, pickable, markShown } from './common.js';

const $ = (id) => document.getElementById(id);
const body = $('watch-body');
const msg = $('watch-msg');
const form = $('watch-form');
let shown = ''; // symbol on the chart
let loaded = false;

function render(symbols) {
  body.replaceChildren(...symbols.map(row));
  markShown(body, shown);
  $('watch-empty').classList.toggle('d-none', symbols.length > 0);
}

async function load() {
  loaded = true;
  hideMessage(msg);
  try {
    render((await api('/api/watchlist')).symbols);
  } catch (err) {
    loaded = false; // try again on the next tab click
    showMessage(msg, err.message);
  }
}

function row(symbol) {
  const tr = document.createElement('tr');
  pickable(tr, symbol);
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'btn btn-sm btn-outline-danger py-0';
  btn.textContent = 'הסר';
  btn.setAttribute('aria-label', `הסר את ${symbol} מהמעקב`);
  btn.addEventListener('click', (e) => {
    e.stopPropagation(); // remove only, do not open the symbol
    remove(symbol, tr, btn);
  });
  const btnCell = document.createElement('td');
  btnCell.className = 'text-end';
  btnCell.append(btn);
  tr.append(cell(symbol, { ltr: true }), btnCell);
  return tr;
}

async function remove(symbol, tr, btn) {
  hideMessage(msg);
  btn.disabled = true;
  try {
    await api(`/api/watchlist/${encodeURIComponent(symbol)}`, { method: 'DELETE' });
    tr.remove();
    $('watch-empty').classList.toggle('d-none', body.children.length > 0);
  } catch (err) {
    btn.disabled = false;
    showMessage(msg, err.message);
  }
}

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  const symbol = form.symbol.value.trim().toUpperCase();
  if (!symbol) return;
  hideMessage(msg);
  try {
    const { symbols } = await api('/api/watchlist', { method: 'POST', body: JSON.stringify({ symbol }) });
    loaded = true;
    render(symbols);
    form.reset();
  } catch (err) {
    showMessage(msg, err.message);
  }
});
$('tab-watch').addEventListener('click', () => {
  if (!loaded) load();
});
document.addEventListener('symbol-loaded', (e) => {
  shown = e.detail;
  markShown(body, shown);
});
