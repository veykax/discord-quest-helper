// State
let allGames = [];
let selectedGame = null;
let isRunning = false;
let statusInterval = null;

const CIRCUMFERENCE = 2 * Math.PI * 82; // 515.22

// Elements
const gamesListEl = document.getElementById("gamesList");
const searchInput = document.getElementById("searchInput");
const selectedGameTitle = document.getElementById("selectedGameTitle");
const selectedGameExe = document.getElementById("selectedGameExe");
const statusBadge = document.getElementById("statusBadge");
const ringFill = document.getElementById("ringFill");
const timerDigits = document.getElementById("timerDigits");
const timerSub = document.getElementById("timerSub");
const timerCaption = document.getElementById("timerCaption");
const startBtn = document.getElementById("startBtn");
const stopBtn = document.getElementById("stopBtn");

const customExeInput = document.getElementById("customExeInput");
const customTitleInput = document.getElementById("customTitleInput");
const customSelectBtn = document.getElementById("customSelectBtn");

// Init
async function init() {
  ringFill.style.strokeDasharray = `${CIRCUMFERENCE}`;
  ringFill.style.strokeDashoffset = `${CIRCUMFERENCE}`;

  await fetchGames();
  setupEventListeners();
  checkInitialStatus();
}

async function fetchGames() {
  try {
    const res = await fetch("/api/games");
    allGames = await res.json();
    renderGames(allGames);
    if (allGames.length > 0) {
      selectGame(allGames[0]);
    }
  } catch (err) {
    console.error("Failed to load games:", err);
  }
}

function renderGames(games) {
  gamesListEl.innerHTML = "";
  games.forEach((game) => {
    const card = document.createElement("div");
    card.className = `game-card ${selectedGame && selectedGame.id === game.id ? "active" : ""}`;
    card.innerHTML = `
      <div class="game-card-left">
        <div class="game-title">${game.title}</div>
        <div class="game-desc">${game.desc}</div>
        <div class="game-exe-pill">${game.exe}</div>
      </div>
      <div class="game-badge">${game.badge || "Игра"}</div>
    `;
    card.onclick = () => selectGame(game);
    gamesListEl.appendChild(card);
  });
}

function selectGame(game) {
  if (isRunning) return;
  selectedGame = game;
  selectedGameTitle.textContent = game.title;
  selectedGameExe.textContent = game.exe;

  document.querySelectorAll(".game-card").forEach((c) => c.classList.remove("active"));
  const activeCard = Array.from(document.querySelectorAll(".game-card")).find(
    (c) => c.querySelector(".game-title").textContent === game.title
  );
  if (activeCard) activeCard.classList.add("active");
}

function setupEventListeners() {
  // Search
  searchInput.addEventListener("input", (e) => {
    const q = e.target.value.toLowerCase().trim();
    const filtered = allGames.filter(
      (g) =>
        g.title.toLowerCase().includes(q) ||
        g.desc.toLowerCase().includes(q) ||
        g.exe.toLowerCase().includes(q)
    );
    renderGames(filtered);
  });

  // Custom game
  customSelectBtn.addEventListener("click", () => {
    const exe = customExeInput.value.trim();
    const title = customTitleInput.value.trim() || "Свой процесс";
    if (!exe) {
      alert("Укажите имя .exe процесса!");
      return;
    }
    const cleanExe = exe.toLowerCase().endsWith(".exe") ? exe : `${exe}.exe`;
    const custom = {
      id: "custom",
      title: title,
      desc: "Пользовательский процесс",
      exe: cleanExe,
      windowTitle: title,
      badge: "Custom",
    };
    selectGame(custom);
  });

  // Buttons
  startBtn.addEventListener("click", startGame);
  stopBtn.addEventListener("click", stopGame);

  // Close sync with backend
  window.addEventListener("beforeunload", () => {
    navigator.sendBeacon("/api/exit");
  });
}

async function startGame() {
  if (!selectedGame || isRunning) return;

  try {
    const res = await fetch("/api/start", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        id: selectedGame.id,
        title: selectedGame.title,
        exe: selectedGame.exe,
        windowTitle: selectedGame.windowTitle || selectedGame.title,
        steamFolder: selectedGame.steamFolder || "",
        steamAppId: selectedGame.steamAppId || "",
      }),
    });

    if (res.ok) {
      setRunningState(true);
      startPolling();
    } else {
      const errText = await res.text();
      alert("Ошибка запуска: " + errText);
    }
  } catch (err) {
    alert("Ошибка связи с сервисом: " + err);
  }
}

async function stopGame() {
  try {
    await fetch("/api/stop", { method: "POST" });
    setRunningState(false);
    resetTimerUI();
  } catch (err) {
    console.error("Stop error:", err);
  }
}

function setRunningState(running) {
  isRunning = running;
  startBtn.disabled = running;
  stopBtn.disabled = !running;

  if (running) {
    statusBadge.textContent = "🟢 Эмуляция активна";
    statusBadge.className = "status-indicator active";
    timerCaption.textContent = `Discord видит игру: ${selectedGame ? selectedGame.exe : ""}`;
    ringFill.classList.add("running");
    timerDigits.classList.add("running");
    timerSub.textContent = "ОСТАЛОСЬ";
  } else {
    statusBadge.textContent = "Готов к запуску";
    statusBadge.className = "status-indicator";
    timerCaption.textContent = "Нажмите «Запустить эмуляцию» для старта";
    ringFill.classList.remove("running");
    timerDigits.classList.remove("running");
    timerSub.textContent = "ТАЙМЕР";
    if (statusInterval) {
      clearInterval(statusInterval);
      statusInterval = null;
    }
  }
}

function startPolling() {
  if (statusInterval) clearInterval(statusInterval);
  statusInterval = setInterval(pollStatus, 1000);
}

async function pollStatus() {
  try {
    const res = await fetch("/api/status");
    const data = await res.json();

    if (!data.running && isRunning) {
      // Completed!
      setRunningState(false);
      onQuestCompleted();
      return;
    }

    if (data.running) {
      updateTimerUI(data.remaining, data.total);
    }
  } catch (err) {
    console.error("Poll error:", err);
  }
}

async function checkInitialStatus() {
  try {
    const res = await fetch("/api/status");
    const data = await res.json();
    if (data.running) {
      if (data.game && data.game.title) {
        selectGame(data.game);
      }
      setRunningState(true);
      updateTimerUI(data.remaining, data.total);
      startPolling();
    }
  } catch (e) {}
}

function updateTimerUI(remaining, total) {
  const mins = Math.floor(remaining / 60);
  const secs = remaining % 60;
  timerDigits.textContent = `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;

  const elapsed = total - remaining;
  const fraction = Math.min(1, Math.max(0, elapsed / total));
  const offset = CIRCUMFERENCE * (1 - fraction);
  ringFill.style.strokeDashoffset = offset;
}

function resetTimerUI() {
  timerDigits.textContent = "15:00";
  ringFill.style.strokeDashoffset = CIRCUMFERENCE;
}

function onQuestCompleted() {
  timerDigits.textContent = "00:00";
  ringFill.style.strokeDashoffset = 0;
  statusBadge.textContent = "🎉 Выполнен!";
  statusBadge.className = "status-indicator completed";
  timerCaption.textContent = "15 минут прошло. Заберите ваши Orbs в Discord!";
  timerSub.textContent = "ВЫПОЛНЕНО";

  playChime();
  alert("🎉 Квест выполнен!\n\n15 минут прошло. Откройте Discord и нажмите «Claim Reward» для получения Orbs!");
}

// Приятный синтетический звук завершения через Web Audio API
function playChime() {
  try {
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const notes = [523.25, 659.25, 783.99, 1046.5]; // C5, E5, G5, C6
    notes.forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = "sine";
      osc.frequency.value = freq;
      gain.gain.setValueAtTime(0.12, ctx.currentTime + i * 0.12);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + i * 0.12 + 0.4);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start(ctx.currentTime + i * 0.12);
      osc.stop(ctx.currentTime + i * 0.12 + 0.45);
    });
  } catch (e) {}
}

init();
