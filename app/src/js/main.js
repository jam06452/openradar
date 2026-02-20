import '../scss/main.scss';

let allLeaks = [];
let currentPage = 1;
let totalPages = 1;
let isLoading = false;
let currentFilter = 'all';
let pollingIntervalId = null;
let tickerQueue = [];
let tickerBusy = false;

function timeAgo(dateString) {
    const seconds = Math.floor((Date.now() - new Date(dateString)) / 1000);
    const intervals = [
        [31536000, 'year'],
        [2592000, 'month'],
        [86400, 'day'],
        [3600, 'hour'],
        [60, 'minute'],
    ];
    for (const [divisor, label] of intervals) {
        const value = Math.floor(seconds / divisor);
        if (value >= 1) return `${value} ${label}${value !== 1 ? 's' : ''} ago`;
    }
    return `${seconds} seconds ago`;
}

function updateIntervalFor(seconds) {
    if (seconds < 60) return 1000;
    if (seconds < 3600) return 60000;
    if (seconds < 86400) return 3600000;
    return 86400000;
}

function startLiveTimestamps() {
    clearTimeout(startLiveTimestamps._t);
    const elements = document.querySelectorAll('[data-detected-at]');
    let next = 86400000;
    elements.forEach(el => {
        const seconds = Math.floor((Date.now() - new Date(el.dataset.detectedAt)) / 1000);
        el.textContent = timeAgo(el.dataset.detectedAt);
        next = Math.min(next, updateIntervalFor(seconds));
    });
    startLiveTimestamps._t = setTimeout(startLiveTimestamps, next);
}

function repoDisplayName(apiUrl) {
    try {
        const path = apiUrl.replace('https://api.github.com/repos/', '');
        const [owner, repo] = path.split('/');
        return { displayName: `${owner}/${repo}`, publicUrl: `https://github.com/${owner}/${repo}` };
    } catch {
        return { displayName: apiUrl, publicUrl: '#' };
    }
}

function animateCounter(el, target, duration) {
    const step = target / (duration / 16);
    let current = 0;
    const tick = () => {
        current = Math.min(current + step, target);
        el.textContent = Math.ceil(current).toLocaleString();
        if (current < target) requestAnimationFrame(tick);
    };
    tick();
}

function createCard(leak) {
    const { displayName, publicUrl } = repoDisplayName(leak.repo_name);
    const fileUrl = `${publicUrl}/blob/main/${leak.file_path}`;
    const card = document.createElement('article');
    card.className = 'card';
    card.innerHTML = `
        <div class="card-header">
            <pre class="card-key"><code>${leak.key}</code></pre>
            <span class="card-provider-badge ${leak.provider}">${leak.provider}</span>
        </div>
        <div class="card-body">
            <div class="card-row card-repo">
                <img src="/git.svg" class="icon" alt="Repo icon" />
                <span>Repo Name:</span>
                <a href="${publicUrl}" target="_blank" rel="noopener noreferrer">${displayName}</a>
            </div>
            <div class="card-row card-filepath">
                <img src="/file.svg" class="icon" alt="File icon" />
                <span>Key path:</span>
                <a href="${fileUrl}" target="_blank" rel="noopener noreferrer" class="path">${leak.file_path}</a>
            </div>
            <div class="card-row card-detected">
                <div class="card-meta-item">
                    <img src="/calendar.svg" class="icon" alt="Calendar icon" />
                    <span>Detected: <strong data-detected-at="${leak.detected_at}">${timeAgo(leak.detected_at)}</strong></span>
                </div>
            </div>
        </div>
    `;
    return card;
}

function clearMessages() {
    document.querySelectorAll('.loading-message, .end-message, .empty-message').forEach(el => el.remove());
}

function showMessage(grid, text, className) {
    clearMessages();
    grid.innerHTML = `<p class="${className} grid-message">${text}</p>`;
}

function appendCards(leaks, prepend = false) {
    const grid = document.getElementById('card-grid');
    if (!grid) return;
    const offset = prepend ? 0 : grid.children.length;
    const fragment = document.createDocumentFragment();
    const ordered = prepend ? [...leaks].reverse() : leaks;
    ordered.forEach((leak, i) => {
        const card = createCard(leak);
        card.style.animationDelay = `${(offset + i) * (prepend ? 100 : 80)}ms`;
        fragment.appendChild(card);
    });
    prepend ? grid.prepend(fragment) : grid.appendChild(fragment);
    startLiveTimestamps();
}

function showEndMessage(grid) {
    const p = document.createElement('p');
    p.className = 'end-message grid-message';
    p.textContent = 'thats all for now (:';
    grid.insertAdjacentElement('afterend', p);
}

async function fetchTotalCount() {
    try {
        const res = await fetch('/api/findings/count');
        if (!res.ok) return;
        const { total_count } = await res.json();
        const el = document.getElementById('leak-count');
        if (el && total_count) animateCounter(el, total_count, 1200);
    } catch { }
}

async function fetchLeaks(page, filter) {
    if (isLoading) return;
    isLoading = true;

    const grid = document.getElementById('card-grid');
    if (page === 1) showMessage(grid, 'Loading findings...', 'loading-message');

    try {
        const provider = filter.toLowerCase() === 'all' ? '*' : filter.toLowerCase();
        const res = await fetch(`/api/findings?page=${page}&page_size=25&provider=${provider}`);

        if (res.status === 404) {
            if (page === 1) showMessage(grid, 'Theres nothing here (yet)! ;)', 'empty-message');
            totalPages = page;
            return;
        }

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const data = await res.json();
        totalPages = data.total_pages || 1;

        if (page === 1) grid.innerHTML = '';

        const findings = data.findings || [];
        appendCards(findings);
        allLeaks = allLeaks.concat(findings);

        if (page === 1 && findings.length === 0) {
            showMessage(grid, 'Theres nothing here (yet)! ;)', 'empty-message');
        }

        if (currentPage >= totalPages && allLeaks.length > 0) showEndMessage(grid);

        if (page === 1) {
            const el = document.getElementById('leak-count');
            if (el && data.total_count) animateCounter(el, data.total_count, 1200);
        }
    } catch {
        showMessage(grid, 'oof cant connect to the server, is it up?', 'empty-message');
    } finally {
        isLoading = false;
    }
}

async function pollForNewLeaks() {
    if (isLoading) return;
    try {
        const provider = currentFilter.toLowerCase() === 'all' ? '*' : currentFilter.toLowerCase();
        const res = await fetch(`/api/findings?page=1&page_size=25&provider=${provider}`);
        if (!res.ok) return;
        const data = await res.json();
        const existingIds = new Set(allLeaks.map(l => l.id));
        const fresh = (data.findings || []).filter(l => !existingIds.has(l.id));
        if (fresh.length > 0) {
            allLeaks.unshift(...fresh);
            appendCards(fresh, true);
        }
    } catch { }
}

function restartPolling() {
    clearInterval(pollingIntervalId);
    pollingIntervalId = setInterval(pollForNewLeaks, 15000);
}

function setupTabFiltering() {
    const tabs = document.querySelector('.tabs');
    if (!tabs) return;
    tabs.addEventListener('click', e => {
        const tab = e.target.closest('.tab');
        if (!tab || isLoading) return;
        clearMessages();
        allLeaks = [];
        currentPage = 1;
        tabs.querySelector('.active')?.classList.remove('active');
        tab.classList.add('active');
        currentFilter = tab.textContent;
        fetchLeaks(currentPage, currentFilter);
        restartPolling();
    });
}

function setupInfiniteScroll() {
    window.addEventListener('scroll', () => {
        const nearBottom = window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 1200;
        if (nearBottom && !isLoading && currentPage < totalPages) {
            currentPage++;
            fetchLeaks(currentPage, currentFilter);
        }
    });
}

function runTicker() {
    if (tickerQueue.length === 0) {
        tickerBusy = false;
        const ticker = document.querySelector('.ticker');
        if (ticker) { ticker.style.animation = 'none'; ticker.innerHTML = ''; }
        return;
    }
    tickerBusy = true;
    const ticker = document.querySelector('.ticker');
    const name = tickerQueue.shift();
    ticker.innerHTML = `<div class="info">Scanning ${name}</div>`;
    ticker.style.animation = 'none';
    ticker.offsetHeight;
    ticker.style.animation = 'scroll-up 2.5s cubic-bezier(0.25, 0.1, 0.25, 1) forwards';
    ticker.addEventListener('animationend', runTicker, { once: true });
}

function addTickerItem(url) {
    try {
        const parts = url.replace('https://api.github.com/repos/', '').split('/');
        tickerQueue.push(`${parts[0]}/${parts[1]}`);
    } catch {
        tickerQueue.push(url);
    }
    if (!tickerBusy) runTicker();
}

function connectWebSocket() {
    const ws = new WebSocket(`wss://${location.host}/ws/live`);
    ws.onmessage = e => {
        try { addTickerItem(JSON.parse(e.data).url); } catch { }
    };
    ws.onclose = () => setTimeout(connectWebSocket, 3000);
    ws.onerror = () => ws.close();
}

document.addEventListener('DOMContentLoaded', () => {
    connectWebSocket();
    fetchLeaks(currentPage, currentFilter).then(() => {
        restartPolling();
        startLiveTimestamps();
    });
    setupTabFiltering();
    setupInfiniteScroll();
});