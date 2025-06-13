const logsContainer = document.getElementById('logs-container');
let lastLogCount = 0; // 이전에 불러온 로그 개수 추적

function fetchLogs() {
    fetch('/api/logs') // Go 서버의 API 엔드포인트 호출
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(logs => {
            // 새로 불러온 로그가 없으면 업데이트하지 않음
            if (logs.length === lastLogCount) {
                return;
            }
            lastLogCount = logs.length;

            // 로딩 메시지 제거
            const loadingMessage = logsContainer.querySelector('.loading-message');
            if (loadingMessage) {
                loadingMessage.remove();
            }

            // 모든 로그를 다시 그리는 대신, 추가된 로그만 효율적으로 추가
            // (간단한 예시에서는 전체를 다시 그리지만, 대규모에서는 최적화 필요)
            logsContainer.innerHTML = ''; // 기존 로그를 모두 지웁니다.
            logs.slice().reverse().forEach(log => { // 최신 로그가 위에 오도록 역순으로 순회
                const logEntryDiv = document.createElement('div');
                logEntryDiv.className = 'log-entry';
                logEntryDiv.innerHTML = `
                    <span class="timestamp">${log.timestamp}</span>
                    <span>${log.file}:</span>
                    <span class="tid">[TID:${log.threadId}]</span>
                    <span class="protocol">[${log.protocol}]</span>
                    <span>Sent <span class="bytes">${log.bytesSent}</span> bytes to</span>
                    <span class="ip">${log.targetIp}</span>
                    <span class="packet">(${log.packetInfo})</span>
                `;
                logsContainer.appendChild(logEntryDiv); // 컨테이너에 추가 (reverse 했으므로 append)
            });
            // 스크롤을 항상 아래로 내립니다 (가장 최신 로그가 보이도록)
            logsContainer.scrollTop = logsContainer.scrollHeight;

        })
        .catch(error => {
            console.error('로그를 불러오는 데 실패했습니다:', error);
            logsContainer.innerHTML = '<p class="error-message">로그를 불러오는 데 문제가 발생했습니다. 서버가 실행 중인지 확인하세요.</p>';
        });
}

// 다크 모드 토글 기능
function toggleDarkMode() {
    document.body.classList.toggle('dark-mode');
}

// 로그 디스플레이 초기화 기능
function clearLogsDisplay() {
    logsContainer.innerHTML = '<p class="loading-message">로그를 기다리는 중...</p>';
    lastLogCount = 0; // 로그 개수 초기화
}

// 1초마다 로그를 새로고침합니다.
setInterval(fetchLogs, 1000);

// 페이지 로드 시 초기 로그를 불러옵니다.
document.addEventListener('DOMContentLoaded', fetchLogs);