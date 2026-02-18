import '../scss/main.scss';

let allLeaks = [];
let currentPage = 1;
let totalPages = 1;
let isLoading = false;
let currentFilter = 'all';
let pollingIntervalId = null;

function timeAgo(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const seconds = Math.floor((now - date) / 1000);

    let interval = seconds / 31536000;
    if (interval > 1) return Math.floor(interval) + " years ago";
    interval = seconds / 2592000;
    if (interval > 1) return Math.floor(interval) + " months ago";
    interval = seconds / 86400;
    if (interval > 1) return Math.floor(interval) + " days ago";
    interval = seconds / 3600;
    if (interval > 1) return Math.floor(interval) + " hours ago";
    interval = seconds / 60;
    if (interval > 1) return Math.floor(interval) + " minutes ago";
    return Math.floor(seconds) + " seconds ago";
}

function getUpdateInterval(seconds) {
    if (seconds < 60) return 1000;
    if (seconds < 3600) return 60000;
    if (seconds < 86400) return 3600000;
    return 86400000;
}

function startLiveTimestamps() {
    function tick() {
        const timeElements = document.querySelectorAll('[data-detected-at]');
        let nextInterval = 86400000;

        timeElements.forEach(el => {
            const dateString = el.getAttribute('data-detected-at');
            const seconds = Math.floor((new Date() - new Date(dateString)) / 1000);
            el.textContent = timeAgo(dateString);
            nextInterval = Math.min(nextInterval, getUpdateInterval(seconds));
        });

        clearTimeout(startLiveTimestamps._timeout);
        startLiveTimestamps._timeout = setTimeout(tick, nextInterval);
    }

    tick();
}

function transformGitHubUrl(apiUrl) {
    const defaultResult = { displayName: apiUrl, publicUrl: '#' };
    if (!apiUrl || typeof apiUrl !== 'string') return defaultResult;
    try {
        if (apiUrl.startsWith('https://api.github.com/repos/')) {
            const pathParts = apiUrl.substring('https://api.github.com/repos/'.length).split('/');
            if (pathParts.length >= 2) {
                const owner = pathParts[0];
                const repo = pathParts[1];
                const displayName = `${owner}/${repo}`;
                const publicUrl = `https://github.com/${owner}/${repo}`;
                return { displayName, publicUrl };
            }
        }
    } catch (e) {
        console.error("Could not parse GitHub API URL:", apiUrl, e);
    }
    return defaultResult;
}

function animateCounter(element, finalValue, duration) {
    let start = 0;
    const increment = finalValue / (duration / 16);
    const updateCount = () => {
        start += increment;
        if (start < finalValue) {
            element.innerText = Math.ceil(start).toLocaleString();
            requestAnimationFrame(updateCount);
        } else {
            element.innerText = finalValue.toLocaleString();
        }
    };
    updateCount();
}

function createCardElement(leak) {
    const card = document.createElement('article');
    card.className = 'card';
    const { displayName, publicUrl } = transformGitHubUrl(leak.repo_name);
    const fileUrl = `${publicUrl}/blob/main/${leak.file_path}`;

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

function prependCards(leaks) {
    const grid = document.getElementById('card-grid');
    if (!grid) return;
    
    leaks.reverse().forEach((leak, index) => {
        const cardElement = createCardElement(leak);
        cardElement.style.animationDelay = `${index * 100}ms`;
        grid.prepend(cardElement);
    });

    startLiveTimestamps();
}

function appendCards(leaks) {
    const grid = document.getElementById('card-grid');
    if (!grid) return;
    const existingCardCount = grid.children.length;
    leaks.forEach((leak, index) => {
        const cardElement = createCardElement(leak);
        cardElement.style.animationDelay = `${(existingCardCount + index) * 80}ms`;
        grid.appendChild(cardElement);
    });

    startLiveTimestamps();
}

function removeMessage(selector) {
    const message = document.querySelector(selector);
    if (message) message.remove();
}

function showMessage(grid, message, className) {
    removeMessage('.loading-message');
    removeMessage('.end-message');
    removeMessage('.empty-message');
    grid.innerHTML = `<p class="${className} grid-message">${message}</p>`;
}

async function fetchTotalCount() {
    try {
        const response = await fetch('/findings/count');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        const leakCountElement = document.getElementById('leak-count');
        if (leakCountElement && data.total_count) {
            animateCounter(leakCountElement, data.total_count, 1200);
        }
    } catch (error) {
        console.error("Failed to fetch total count:", error);
    }
}

async function fetchLeaks(page, filter) {
    if (isLoading) return;
    isLoading = true;

    const grid = document.getElementById('card-grid');
    if (page === 1) {
        showMessage(grid, 'Loading findings...', 'loading-message');
    }

    try {
        const provider = filter.toLowerCase() === 'all' ? '*' : filter.toLowerCase();
        const response = await fetch(`/findings?page=${page}&page_size=25&provider=${provider}`);
        
        if (response.status === 404) {
            if (page === 1) {
                showMessage(grid, 'Theres nothing here (yet)! ;)', 'empty-message');
            }
            totalPages = page;
            return;
        }

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        totalPages = data.total_pages || 1;

        if (page === 1) {
            grid.innerHTML = '';
        }
        
        appendCards(data.findings || []);
        allLeaks = allLeaks.concat(data.findings || []);

        if (page === 1 && (!data.findings || data.findings.length === 0)) {
            showMessage(grid, 'Theres nothing here (yet)! ;)', 'empty-message');
        }

        if (currentPage >= totalPages && allLeaks.length > 0) {
            const endMessage = document.createElement('p');
            endMessage.className = 'end-message grid-message';
            endMessage.textContent = 'thats all for now (:';
            grid.insertAdjacentElement('afterend', endMessage);
        }

        if (page === 1) {
            const leakCountElement = document.getElementById('leak-count');
            if (leakCountElement && data.total_count) {
                animateCounter(leakCountElement, data.total_count, 1200);
            }
        }

    } catch (error) {
        console.error("Failed to fetch findings:", error);
        showMessage(grid, 'oof cant connect to the server, is it up?', 'empty-message');
    } finally {
        isLoading = false;
    }
}

async function pollForNewLeaks() {
    if (isLoading) return;
    try {
        const provider = currentFilter.toLowerCase() === 'all' ? '*' : currentFilter.toLowerCase();
        const response = await fetch(`/findings?page=1&page_size=25&provider=${provider}`);
        if (!response.ok) return;

        const data = await response.json();
        const newFindings = data.findings || [];
        
        const existingIds = new Set(allLeaks.map(leak => leak.id));
        const uniqueNewFindings = newFindings.filter(leak => !existingIds.has(leak.id));

        if (uniqueNewFindings.length > 0) {
            allLeaks.unshift(...uniqueNewFindings);
            prependCards(uniqueNewFindings);
        }

    } catch (error) {
        console.error("Polling failed:", error);
    }
}

function setupTabFiltering() {
    const tabsContainer = document.querySelector('.tabs');
    if (!tabsContainer) return;

    tabsContainer.addEventListener('click', (event) => {
        const clickedTab = event.target.closest('.tab');
        if (!clickedTab || isLoading) return;

        removeMessage('.end-message');
        allLeaks = [];
        currentPage = 1;
        
        const currentActiveTab = tabsContainer.querySelector('.active');
        if (currentActiveTab) currentActiveTab.classList.remove('active');
        clickedTab.classList.add('active');

        currentFilter = clickedTab.textContent;
        fetchLeaks(currentPage, currentFilter);

        if (pollingIntervalId) clearInterval(pollingIntervalId);
        pollingIntervalId = setInterval(pollForNewLeaks, 15000);
    });
}

function setupInfiniteScroll() {
    window.addEventListener('scroll', () => {
        const scrolledTo = window.innerHeight + window.scrollY;
        const totalHeight = document.documentElement.scrollHeight;
        const isNearBottom = scrolledTo >= totalHeight - 1200;
        
        if (isNearBottom && !isLoading && currentPage < totalPages) {
            currentPage++;
            fetchLeaks(currentPage, currentFilter);
        }
    });
}

document.addEventListener('DOMContentLoaded', () => {
    fetchLeaks(currentPage, currentFilter).then(() => {
        if (pollingIntervalId) clearInterval(pollingIntervalId);
        pollingIntervalId = setInterval(pollForNewLeaks, 15000);
        startLiveTimestamps();
    });
    setupTabFiltering();
    setupInfiniteScroll();
});