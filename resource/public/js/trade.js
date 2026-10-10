// Paper trading: account summary, order form, positions, open orders.
import { api, usd, num, dateTime, showMessage, hideMessage, cell } from './common.js';

const $ = (id) => document.getElementById(id);
const orderForm = $('order-form');
const orderMsg = $('order-msg');
const tradeMsg = $('trade-msg');

const SIDE = { buy: 'קנייה', sell: 'מכירה' };
const TYPE = { market: 'שוק', limit: 'לימיט' };

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
  for (const p of positions) {
    const pl = Number(p.unrealizedPl);
    const tr = document.createElement('tr');
    tr.append(
      cell(p.symbol, { ltr: true }),
      cell(num(p.qty), { ltr: true }),
      cell(usd(p.avgEntryPrice), { ltr: true }),
      cell(usd(p.currentPrice), { ltr: true }),
      cell(usd(p.marketValue), { ltr: true }),
      cell(usd(p.unrealizedPl), { ltr: true, cls: pl >= 0 ? 'text-up' : 'text-down' }),
    );
    body.append(tr);
  }
  $('positions-empty').classList.toggle('d-none', positions.length > 0);
}

async function loadOrders() {
  const { orders } = await api('/api/orders');
  const body = $('orders-body');
  body.replaceChildren();
  for (const o of orders) {
    const tr = document.createElement('tr');
    const btnCell = document.createElement('td');
    const btn = document.createElement('button');
    btn.className = 'btn btn-sm btn-outline-danger';
    btn.textContent = 'ביטול';
    btn.addEventListener('click', () => cancelOrder(o.id, btn));
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
  $('orders-empty').classList.toggle('d-none', orders.length > 0);
}

async function refresh() {
  hideMessage(tradeMsg);
  try {
    await Promise.all([loadAccount(), loadPositions(), loadOrders()]);
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
});

$('refresh-btn').addEventListener('click', refresh);
refresh();
