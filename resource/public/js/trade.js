// Paper trading: account summary, order form, positions, open orders.
// After each refresh, 'trades-changed' gives the positions and orders to the chart (chart-trades.js).
import { api, usd, num, dateTime, showMessage, hideMessage, cell, pickable, markShown } from './common.js';

const $ = (id) => document.getElementById(id);
const orderForm = $('order-form');
const orderMsg = $('order-msg');
const tradeMsg = $('trade-msg');

const SIDE = { buy: 'קנייה', sell: 'מכירה' };
const TYPE = { market: 'שוק', limit: 'לימיט' };
let shown = ''; // symbol on the chart, highlighted in the tables
const selected = new Set(); // position symbols checked for "close selected"
const allBox = $('positions-all');
const closeBtn = $('close-selected');

// Sync the "select all" box and the close button with the checked rows.
function syncSelection() {
  const boxes = [...$('positions-body').querySelectorAll('input[type=checkbox]')];
  const n = boxes.filter((b) => b.checked).length;
  allBox.checked = n > 0 && n === boxes.length;
  allBox.indeterminate = n > 0 && n < boxes.length;
  closeBtn.disabled = n === 0;
  closeBtn.textContent = n ? `סגור ${n} נבחרות במחיר שוק` : 'סגור נבחרות במחיר שוק';
}

async function loadAccount() {
  const a = await api('/api/account');
  $('acc-equity').textContent = usd(a.equity);
  $('acc-cash').textContent = usd(a.cash);
  $('acc-buying').textContent = usd(a.buyingPower);
}

async function loadPositions() {
  const { positions } = await api('/api/positions');
  const body = $('positions-body');
  body.replaceChildren();
  let total = 0;
  const open = new Set(positions.map((p) => p.symbol));
  for (const s of selected) if (!open.has(s)) selected.delete(s); // drop closed positions
  for (const p of positions) {
    const pl = Number(p.unrealizedPl);
    total += pl;
    const tr = document.createElement('tr');
    pickable(tr, p.symbol);
    const boxCell = document.createElement('td');
    const box = document.createElement('input');
    box.type = 'checkbox';
    box.className = 'form-check-input';
    box.checked = selected.has(p.symbol);
    box.setAttribute('aria-label', `בחר ${p.symbol}`);
    box.addEventListener('click', (e) => e.stopPropagation()); // select only, do not open the symbol
    box.addEventListener('change', () => {
      if (box.checked) selected.add(p.symbol);
      else selected.delete(p.symbol);
      syncSelection();
    });
    boxCell.append(box);
    tr.append(
      boxCell,
      cell(p.symbol, { ltr: true }),
      cell(num(p.qty), { ltr: true }),
      cell(usd(p.avgEntryPrice), { ltr: true }),
      cell(usd(p.currentPrice), { ltr: true }),
      cell(usd(p.marketValue), { ltr: true }),
      cell(usd(p.unrealizedPl), { ltr: true, cls: pl >= 0 ? 'text-up' : 'text-down' }),
    );
    body.append(tr);
  }
  markShown(body, shown);
  const sum = $('positions-total');
  sum.textContent = usd(total);
  sum.className = total >= 0 ? 'text-up' : 'text-down';
  $('positions-foot').classList.toggle('d-none', positions.length === 0);
  $('positions-empty').classList.toggle('d-none', positions.length > 0);
  closeBtn.classList.toggle('d-none', positions.length === 0);
  allBox.disabled = positions.length === 0;
  syncSelection();
  return positions;
}

async function loadOrders() {
  const { orders } = await api('/api/orders');
  const body = $('orders-body');
  body.replaceChildren();
  for (const o of orders) {
    const tr = document.createElement('tr');
    pickable(tr, o.symbol);
    const btnCell = document.createElement('td');
    const btn = document.createElement('button');
    btn.className = 'btn btn-sm btn-outline-danger';
    btn.textContent = 'ביטול';
    btn.addEventListener('click', (e) => {
      e.stopPropagation(); // cancel only, do not open the symbol
      cancelOrder(o.id, btn);
    });
    btnCell.append(btn);
    tr.append(
      cell(dateTime(o.createdAt)),
      cell(o.symbol, { ltr: true }),
      cell(SIDE[o.side] || o.side),
      cell(TYPE[o.type] || o.type),
      cell(num(o.qty), { ltr: true }),
      cell(o.limitPrice ? usd(o.limitPrice) : '-', { ltr: true }),
      cell(o.status, { ltr: true }),
      btnCell,
    );
    body.append(tr);
  }
  markShown(body, shown);
  $('orders-empty').classList.toggle('d-none', orders.length > 0);
  return orders;
}

async function refresh() {
  hideMessage(tradeMsg);
  try {
    const [, positions, orders] = await Promise.all([loadAccount(), loadPositions(), loadOrders()]);
    document.dispatchEvent(new CustomEvent('trades-changed', { detail: { positions, orders } })); // chart lines
  } catch (err) {
    showMessage(tradeMsg, err.message);
  }
}

async function cancelOrder(id, btn) {
  btn.disabled = true;
  try {
    await api(`/api/orders/${encodeURIComponent(id)}`, { method: 'DELETE' });
    await refresh();
  } catch (err) {
    showMessage(tradeMsg, err.message);
    btn.disabled = false;
  }
}

allBox.addEventListener('change', () => {
  const on = allBox.checked; // read once: syncSelection() changes allBox during the loop
  for (const box of $('positions-body').querySelectorAll('input[type=checkbox]')) {
    box.checked = on;
    box.dispatchEvent(new Event('change'));
  }
});

// Close each selected position with a market order. One failure does not stop the others.
closeBtn.addEventListener('click', async () => {
  const symbols = [...selected];
  if (!symbols.length || !confirm(`לסגור במחיר שוק: ${symbols.join(', ')}?`)) return;
  hideMessage(tradeMsg);
  closeBtn.disabled = true;
  const results = await Promise.allSettled(
    symbols.map((s) => api(`/api/positions/${encodeURIComponent(s)}`, { method: 'DELETE' })),
  );
  const errors = results.filter((r) => r.status === 'rejected').map((r) => r.reason.message);
  await refresh();
  if (errors.length) showMessage(tradeMsg, errors.join('\n'));
});

orderForm.type.addEventListener('change', () => {
  const limit = orderForm.type.value === 'limit';
  $('limit-group').classList.toggle('d-none', !limit);
  orderForm.limitPrice.required = limit;
});

orderForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  hideMessage(orderMsg);
  const f = orderForm;
  const btn = f.querySelector('button[type=submit]');
  btn.disabled = true;
  try {
    const o = await api('/api/orders', {
      method: 'POST',
      body: JSON.stringify({
        symbol: f.symbol.value.trim().toUpperCase(),
        side: f.side.value,
        type: f.type.value,
        qty: f.qty.value,
        limitPrice: f.type.value === 'limit' ? f.limitPrice.value : '',
      }),
    });
    showMessage(orderMsg, `הפקודה נשלחה: ${SIDE[o.side]} ${num(o.qty)} ${o.symbol} (${o.status})`, 'success');
    await refresh();
  } catch (err) {
    showMessage(orderMsg, err.message);
  } finally {
    btn.disabled = false;
  }
});

// The order symbol follows the chart; it can still be edited before sending.
document.addEventListener('symbol-loaded', (e) => {
  orderForm.symbol.value = e.detail;
  shown = e.detail;
  markShown($('positions-body'), shown);
  markShown($('orders-body'), shown);
});

$('refresh-btn').addEventListener('click', refresh);
refresh();
