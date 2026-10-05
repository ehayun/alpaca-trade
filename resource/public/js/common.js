// Shared helpers: API calls and Hebrew/US formatting.

// api calls a JSON endpoint wrapped by GoFrame's {code, message, data} envelope.
export async function api(path, options = {}) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  let body;
  try {
    body = await res.json();
  } catch {
    throw new Error('שגיאת שרת, נסו שוב');
  }
  if (body.code !== 0) throw new Error(body.message || 'שגיאה לא צפויה');
  return body.data;
}

const usdFmt = new Intl.NumberFormat('he-IL', { style: 'currency', currency: 'USD' });
const numFmt = new Intl.NumberFormat('he-IL', { maximumFractionDigits: 4 });
const dateTimeFmt = new Intl.DateTimeFormat('he-IL', {
  timeZone: 'Asia/Jerusalem',
  day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit',
});

export const usd = (v) => (v === null || v === undefined || v === '' ? '-' : usdFmt.format(Number(v)));
export const num = (v) => (v === null || v === undefined || v === '' ? '-' : numFmt.format(Number(v)));
export const dateTime = (s) => dateTimeFmt.format(new Date(s));

// showMessage puts text into a Bootstrap alert element and reveals it.
export function showMessage(el, text, kind = 'danger') {
  el.className = `alert alert-${kind}`;
  el.textContent = text;
}

export function hideMessage(el) {
  el.className = 'alert d-none';
  el.textContent = '';
}

// cell builds a <td>; numbers and symbols are isolated so they don't flip inside RTL text.
export function cell(text, { ltr = false, cls = '' } = {}) {
  const td = document.createElement('td');
  if (cls) td.className = cls;
  if (ltr) {
    const b = document.createElement('bdi');
    b.textContent = text;
    td.append(b);
  } else {
    td.textContent = text;
  }
  return td;
}
