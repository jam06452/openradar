import '../scss/main.scss';
import '../scss/leaderboard.scss';

function setLeaderboard() {
    fetch('/api/leaderboard')
        .then(response => response.json())
        .then(data => {
            const leaderboard = document.getElementById('leaderboard');
            leaderboard.innerHTML = '';
            data.forEach((user, index) => {
                const rank = index + 1;
                const leaderboardStat = document.createElement('div');
                leaderboardStat.classList.add('leader-stat');
                let hatHtml = '';
                if (rank === 1) {
                    hatHtml = `<img src="/jester.svg" class="jester-hat">`;
                }
                leaderboardStat.innerHTML = `
                    ${hatHtml}
                    <img src="${user.avatar}" class="leaderboard-profile">
                    <div class="leaderboard-inside">
                        <h1>#${rank}</h1>
                    </div>
                    <h1>@${user.username}</h1>
                    <span class="leak-count">${user.leaks} leaks</span>
                `;
                leaderboard.appendChild(leaderboardStat);
            });
        })
        .catch(error => console.error(error));
}

function setStats() {
    fetch('/api/findings/count')
        .then(response => response.json())
        .then(data => {
            const leaksEl = document.getElementById('stat-leaks');
            if (leaksEl) {
                leaksEl.textContent = data.total_count.toLocaleString();
            }
        })
        .catch(error => console.error(error));

    fetch('/api/repositories/count')
        .then(response => response.json())
        .then(data => {
            const reposEl = document.getElementById('stat-repos');
            if (reposEl) {
                reposEl.textContent = data.total_count.toLocaleString();
            }
        })
        .catch(error => console.error(error));
}

document.addEventListener('DOMContentLoaded', () => {
    setLeaderboard();
    setStats();
});