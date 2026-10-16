// EMA scanner side panel: choose the spread limit and whether the price must be close too, start a scan,
// show progress, list the matches with live prices; a click loads the chart and info.
// On open, a saved scan (if any) is offered instead of a new scan.
import { api, showMessage, hideMessage, num, usd, dateTime, pickable, markShown } from './common.js';

const $ = (id) => document.getElementById(id);
const btn = $('scan-btn');
const statusEl = $('scan-status');
const msg = $('scan-msg');
const list = $('scan-list');
const saved = $('scan-saved');
const pctInput = $('scan-pct');
const priceClose = $('scan-price-close');
const allBox = $('scan-all');
const buyBtn = $('scan-buy-selected');

const POLL_MS = 2000;
const DEFAULT_QTY = 10;
let shown = ''; // symbol on the chart, highlighted in the list

// The settings a scan ran with, e.g. "עד 2% כולל המחיר".
function settingsText(s) {
  return `עד ${num(s.maxSpreadPct)}%${s.priceClose ? ' כולל המחיר' : ''}`;
}

function statusText(s) {
  if (s.running) return `סורק… ${num(s.done)} מתוך ${num(s.total)} מניות (${settingsText(s)})`;
  if (!s.finishedAt) return 'עדיין לא נסרק.';
  return `נסרקו ${num(s.scanned)} מניות · ${dateTime(s.finishedAt)} · נמצאו ${num(s.matches.length)} (${settingsText(s)})`;
}

function td(text, cls = '') {
  const cell = document.createElement('td');
  cell.className = cls;
  const b = document.createElement('bdi');
  b.textContent = text;
  cell.append(b);
  return cell;
}

// The spread a scan used: with the close as a fourth value when the price rule was on.
function row(m, withPrice) {
  const tr = document.createElement('tr');
  tr.title = `EMA 20: ${m.ema20.toFixed(2)} · EMA 100: ${m.ema100.toFixed(2)} · EMA 200: ${m.ema200.toFixed(2)}`
    + ` · סגירה: ${m.close.toFixed(2)}`;
  const stop = (el) => el.addEventListener('click', (e) => e.stopPropagation()); // do not open the symbol

  const box = document.createElement('input');
  box.type = 'checkbox';
  box.className = 'form-check-input scan-select';
  box.setAttribute('aria-label', `בחר ${m.symbol}`);
  box.addEventListener('change', syncSelection);
  stop(box);

  const qty = document.createElement('input');
  qty.type = 'number';
  qty.className = 'form-control form-control-sm scan-qty';
  qty.dir = 'ltr';
  qty.min = '0';
  qty.step = 'any';
  qty.value = DEFAULT_QTY;
  qty.setAttribute('aria-label', `כמות ${m.symbol}`);
  stop(qty);

  const buy = document.createElement('button');
  buy.type = 'button';
  buy.className = 'btn btn-sm btn-outline-success';
  buy.textContent = 'קנה';
  buy.addEventListener('click', async (e) => {
    e.stopPropagation();
    buy.disabled = true;
    await buyAll([tr]);
    buy.disabled = false;
  });

  const wrap = (el) => {
    const cell = document.createElement('td');
    cell.append(el);
    return cell;
  };
  tr.append(
    wrap(box),
    td(m.symbol, 'fw-medium'),
    wrap(qty),
    wrap(buy),
    td(usd(m.close), 'text-end scan-price'), // close at scan time until the live price arrives
    td(`${(withPrice ? m.priceSpreadPct : m.spreadPct).toFixed(2)}%`, 'text-end text-body-secondary'),
  );
  pickable(tr, m.symbol);
  return tr;
}

function highlight() {
  markShown(list, shown);
}

const boxes = () => [...list.querySelectorAll('.scan-select')];

// Sync the "select all" box and the buy button with the checked rows.
function syncSelection() {
  const all = boxes();
  const n = all.filter((b) => b.checked).length;
  allBox.checked = n > 0 && n === all.length;
  allBox.indeterminate = n > 0 && n < all.length;
  buyBtn.disabled = n === 0;
  buyBtn.textContent = n ? `קנה ${n} נבחרות במחיר שוק` : 'קנה נבחרות במחיר שוק';
}

// Send a market buy (day order) for each row, with the quantity in its field.
// One failure does not stop the others; the result of all orders shows in one message.
async function buyAll(rows) {
  hideMessage(msg);
  const results = await Promise.allSettled(rows.map((tr) => api('/api/orders', {
    method: 'POST',
    body: JSON.stringify({
      symbol: tr.dataset.symbol, side: 'buy', type: 'market', qty: tr.querySelector('.scan-qty').value,
    }),
  })));
  const ok = results.filter((r) => r.status === 'fulfilled').map((r) => `${num(r.value.qty)} ${r.value.symbol}`);
  const errors = results.filter((r) => r.status === 'rejected').map((r) => r.reason.message);
  if (ok.length) document.dispatchEvent(new Event('orders-changed')); // trade.js reloads positions and orders
  const lines = [];
  if (ok.length) lines.push(`נשלחה קנייה במחיר שוק: ${ok.join(', ')}`);
  lines.push(...errors);
  showMessage(msg, lines.join('\n'), errors.length ? 'danger' : 'success');
}

allBox.addEventListener('change', () => {
  for (const b of boxes()) b.checked = allBox.checked;
  syncSelection();
});

buyBtn.addEventListener('click', async () => {
  const rows = boxes().filter((b) => b.checked).map((b) => b.closest('tr'));
  const summary = rows.map((tr) => `${tr.querySelector('.scan-qty').value} ${tr.dataset.symbol}`).join(', ');
  if (!rows.length || !confirm(`לקנות במחיר שוק: ${summary}?`)) return;
  buyBtn.disabled = true;
  await buyAll(rows);
  syncSelection();
});

// Live prices for the listed symbols, in one request. On failure the scan-time close stays.
async function refreshPrices(matches) {
  if (!matches.length) return;
  try {
    const { prices } = await api(`/api/prices?symbols=${encodeURIComponent(matches.map((m) => m.symbol).join(','))}`);
    for (const tr of list.children) {
      const p = prices[tr.dataset.symbol];
      if (p !== undefined) tr.querySelector('.scan-price bdi').textContent = usd(p);
    }
  } catch (err) {
    console.warn('live prices:', err.message);
  }
}

function render(s) {
  btn.disabled = s.running;
  statusEl.textContent = statusText(s);
  if (s.failed) showMessage(msg, 'הסריקה נכשלה, נסו שוב');
  else hideMessage(msg);
  if (!s.running) {
    list.replaceChildren(...s.matches.map((m) => row(m, s.priceClose)));
    $('scan-table').hidden = s.matches.length === 0;
    buyBtn.hidden = s.matches.length === 0;
    syncSelection();
    highlight();
    refreshPrices(s.matches);
  }
}

// Poll until the scan finishes; the server keeps scanning even if the page is closed.
async function follow(s) {
  render(s);
  while (s.running) {
    await new Promise((r) => setTimeout(r, POLL_MS));
    try {
      s = await api('/api/scan');
    } catch (err) {
      showMessage(msg, err.message);
      btn.disabled = false;
      return;
    }
    render(s);
  }
}

async function startScan() {
  if (!pctInput.checkValidity()) {
    showMessage(msg, 'הפער חייב להיות בין 0.1% ל-20%');
    pctInput.focus();
    return;
  }
  saved.hidden = true;
  btn.disabled = true;
  const body = {
    maxSpreadPct: Number(pctInput.value),
    priceClose: priceClose.checked,
  };
  try {
    follow(await api('/api/scan', { method: 'POST', body: JSON.stringify(body) }));
  } catch (err) {
    showMessage(msg, err.message);
    btn.disabled = false;
  }
}

btn.addEventListener('click', startScan);
$('scan-saved-rescan').addEventListener('click', startScan);

document.addEventListener('symbol-loaded', (e) => {
  shown = e.detail;
  highlight();
});

// On open: follow a running scan; otherwise offer the saved scan, if there is one.
async function init() {
  const s = await api('/api/scan');
  if (s.running) return follow(s);
  const last = await api('/api/scan/saved');
  if (!last) return render(s);
  $('scan-saved-text').textContent =
    `נמצאה סריקה שמורה מ-${dateTime(last.finishedAt)} (${num(last.matches.length)} מניות, ${settingsText(last)}).`;
  $('scan-saved-use').onclick = () => {
    saved.hidden = true;
    render(last);
  };
  statusEl.textContent = '';
  saved.hidden = false;
}

init().catch((err) => showMessage(msg, err.message));
