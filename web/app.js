// State
let allGames = [];
let selectedGame = null;
let isRunning = false;
let statusInterval = null;
let isSoundEnabled = localStorage.getItem("dqh_sound") !== "false";

const CIRCUMFERENCE = 2 * Math.PI * 82; // 515.22

// Elements
const gamesListEl = document.getElementById("gamesList");
const gamesCountEl = document.getElementById("gamesCount");
const searchInput = document.getElementById("searchInput");
const selectedGameTitle = document.getElementById("selectedGameTitle");
const selectedGameExe = document.getElementById("selectedGameExe");
const statusBadge = document.getElementById("statusBadge");
const statusText = document.getElementById("statusText");
const ringFill = document.getElementById("ringFill");
const timerDigits = document.getElementById("timerDigits");
const timerSub = document.getElementById("timerSub");
const timerPercent = document.getElementById("timerPercent");
const timerCaption = document.getElementById("timerCaption");
const startBtn = document.getElementById("startBtn");
const stopBtn = document.getElementById("stopBtn");

const customExeInput = document.getElementById("customExeInput");
const customTitleInput = document.getElementById("customTitleInput");
const customSelectBtn = document.getElementById("customSelectBtn");

const soundToggleBtn = document.getElementById("soundToggleBtn");
const soundOnIcon = document.getElementById("soundOnIcon");
const soundOffIcon = document.getElementById("soundOffIcon");

const toastEl = document.getElementById("toast");
const toastCloseBtn = document.getElementById("toastCloseBtn");
const toastTitle = document.getElementById("toastTitle");
const toastMsg = document.getElementById("toastMsg");

// Init
async function init() {
  ringFill.style.strokeDasharray = `${CIRCUMFERENCE}`;
  ringFill.style.strokeDashoffset = `${CIRCUMFERENCE}`;

  updateSoundUI();
  await fetchGames();
  setupEventListeners();
  checkInitialStatus();
}

function updateSoundUI() {
  if (isSoundEnabled) {
    soundOnIcon.classList.remove("hidden");
    soundOffIcon.classList.add("hidden");
  } else {
    soundOnIcon.classList.add("hidden");
    soundOffIcon.classList.remove("hidden");
  }
}

// Ограничение минимального размера окна
const MIN_WINDOW_WIDTH = 840;
const MIN_WINDOW_HEIGHT = 580;

window.addEventListener("resize", () => {
  if (window.outerWidth < MIN_WINDOW_WIDTH || window.outerHeight < MIN_WINDOW_HEIGHT) {
    const targetW = Math.max(window.outerWidth, MIN_WINDOW_WIDTH);
    const targetH = Math.max(window.outerHeight, MIN_WINDOW_HEIGHT);
    window.resizeTo(targetW, targetH);
  }
});

function loadCustomGames() {
  try {
    const raw = localStorage.getItem("dqh_custom_games");
    return raw ? JSON.parse(raw) : [];
  } catch (e) {
    return [];
  }
}

function saveCustomGames(games) {
  try {
    localStorage.setItem("dqh_custom_games", JSON.stringify(games));
  } catch (e) {}
}

async function fetchGames() {
  try {
    const res = await fetch("/api/games");
    const serverGames = await res.json();
    const savedCustom = loadCustomGames();
    allGames = [...serverGames, ...savedCustom];
    gamesCountEl.textContent = allGames.length;
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
  const cards = Array.from(document.querySelectorAll(".game-card"));
  const activeCard = cards.find(
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

  // Sound Toggle
  soundToggleBtn.addEventListener("click", () => {
    isSoundEnabled = !isSoundEnabled;
    localStorage.setItem("dqh_sound", isSoundEnabled);
    updateSoundUI();
  });

  // Custom game
  customSelectBtn.addEventListener("click", () => {
    const exe = customExeInput.value.trim();
    const title = customTitleInput.value.trim() || "Свой процесс";
    if (!exe) {
      showToast("Ошибка ввода", "Укажите имя исполняемого файла процесса (.exe)");
      return;
    }
    const cleanExe = exe.toLowerCase().endsWith(".exe") ? exe : `${exe}.exe`;
    const custom = {
      id: "custom_" + Date.now(),
      title: title,
      desc: "Пользовательский процесс",
      exe: cleanExe,
      windowTitle: title,
      badge: "Custom",
    };
    
    // Сохраняем в localStorage, чтобы игра не исчезала после перезапуска
    const saved = loadCustomGames();
    saved.push(custom);
    saveCustomGames(saved);

    allGames.push(custom);
    gamesCountEl.textContent = allGames.length;
    renderGames(allGames);
    selectGame(custom);
    customExeInput.value = "";
    customTitleInput.value = "";
    showToast("Процесс сохранён", `Процесс ${cleanExe} добавлен в ваш список`);
  });

  // Быстрый поиск по нажатию клавиши '/'
  window.addEventListener("keydown", (e) => {
    if (e.key === "/" && document.activeElement !== searchInput && document.activeElement !== customExeInput && document.activeElement !== customTitleInput) {
      e.preventDefault();
      searchInput.focus();
    }
  });

  // Buttons
  startBtn.addEventListener("click", startGame);
  stopBtn.addEventListener("click", stopGame);

  // Toast close
  toastCloseBtn.addEventListener("click", () => {
    toastEl.classList.add("hidden");
  });

  // Открыть Discord через локальный бэкенд (без браузерной навигации)
  const openDiscord = async () => {
    try {
      await fetch("/api/open-discord", { method: "POST" });
    } catch (e) {
      console.error("Open discord error:", e);
    }
  };

  const openDiscordBtn = document.getElementById("openDiscordBtn");
  if (openDiscordBtn) openDiscordBtn.addEventListener("click", openDiscord);
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
      showToast("Ошибка запуска", errText);
    }
  } catch (err) {
    showToast("Ошибка связи", "Не удалось связаться с локальным сервисом: " + err);
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
    statusBadge.className = "status-indicator active";
    statusText.textContent = "Эмуляция активна";
    timerCaption.textContent = `Discord видит процесс: ${selectedGame ? selectedGame.exe : ""}`;
    ringFill.classList.add("running");
    timerDigits.classList.add("running");
    timerSub.textContent = "ОСТАЛОСЬ";
  } else {
    statusBadge.className = "status-indicator";
    statusText.textContent = "В ожидании";
    timerCaption.textContent = "Выберите процесс и нажмите «Запустить эмуляцию»";
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
      setRunningState(false);
      if (data.remaining <= 2) {
        onQuestCompleted();
      } else {
        resetTimerUI();
        showToast("Эмуляция остановлена", "Окно игры было закрыто пользователем");
      }
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
  const percent = Math.round(fraction * 100);
  timerPercent.textContent = `${percent}%`;

  const offset = CIRCUMFERENCE * (1 - fraction);
  ringFill.style.strokeDashoffset = offset;
}

function resetTimerUI() {
  timerDigits.textContent = "15:00";
  timerPercent.textContent = "0%";
  ringFill.style.strokeDashoffset = CIRCUMFERENCE;
}

function onQuestCompleted() {
  timerDigits.textContent = "00:00";
  timerPercent.textContent = "100%";
  ringFill.style.strokeDashoffset = 0;
  statusBadge.className = "status-indicator completed";
  statusText.textContent = "Выполнено";
  timerCaption.textContent = "15 минут истекли. Награда готова к получению в Discord!";
  timerSub.textContent = "РЕЗУЛЬТАТ";

  if (isSoundEnabled) {
    playChime();
  }

  showToast("Квест завершен!", "15 минут успешно завершены. Откройте Discord для получения награды.");
}

function showToast(title, message) {
  toastTitle.textContent = title;
  toastMsg.textContent = message;
  toastEl.classList.remove("hidden");
  setTimeout(() => {
    toastEl.classList.add("hidden");
  }, 7000);
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
