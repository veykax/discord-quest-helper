// Localization Dictionaries
const i18n = {
  en: {
    langCode: "EN",
    questProcesses: "Processes",
    searchPlaceholder: "Search game or .exe...",
    customProcess: "Custom Process",
    selectBtn: "Add",
    openDiscord: "Open Discord",
    statusIdle: "Idle",
    statusRunning: "Running",
    statusCompleted: "Completed",
    captionIdle: "Select a process and start emulation",
    captionRunning: (exe) => `Emulating process: ${exe}`,
    captionCompleted: "Quest completed. Claim reward in Discord.",
    startBtn: "Start (15 min)",
    startBtnRem: (rem) => `Start (${rem})`,
    stopBtn: "Stop",
    questBadgeReady: "Ready",
    questBadgeClaimed: "Claimed",
    smartPreFillHint: (rem) => `Smart timer: ${rem}`,
    readyToClaimHint: "Reward ready in Discord",
    cacheStatusSynced: (tStr) => `Discord cache: ${tStr}`,
    cacheStatusNotFound: "Discord cache not found",
    toastCompleteTitle: "Quest Completed",
    toastCompleteMsg: "Time elapsed. Open Discord to claim your reward.",
    toastStoppedTitle: "Stopped",
    toastStoppedMsg: "Game simulation window was closed by the user",
    toastInputErrTitle: "Input Error",
    toastInputErrMsg: "Please specify an executable name (.exe)",
    toastSavedTitle: "Saved",
    toastSavedMsg: (exe) => `Process ${exe} saved to list`,
    toastLaunchErrTitle: "Launch Error",
    toastConnErrTitle: "Connection Error",
    toastConnErrMsg: "Failed to communicate with service: ",
    soundTitle: "Completion sound",
  },
  ru: {
    langCode: "RU",
    questProcesses: "Процессы",
    searchPlaceholder: "Поиск игры или .exe...",
    customProcess: "Свой процесс",
    selectBtn: "Добавить",
    openDiscord: "В Discord",
    statusIdle: "В ожидании",
    statusRunning: "Активен",
    statusCompleted: "Завершено",
    captionIdle: "Выберите процесс и запустите эмуляцию",
    captionRunning: (exe) => `Процесс эмуляции: ${exe}`,
    captionCompleted: "Квест завершен. Награда готова в Discord.",
    startBtn: "Запустить (15 мин)",
    startBtnRem: (rem) => `Запустить (${rem})`,
    stopBtn: "Остановить",
    questBadgeReady: "Готов",
    questBadgeClaimed: "Получен",
    smartPreFillHint: (rem) => `Авто-таймер: ${rem}`,
    readyToClaimHint: "Награда ждет в Discord",
    cacheStatusSynced: (tStr) => `Кэш Discord: ${tStr}`,
    cacheStatusNotFound: "Кэш Discord не найден",
    toastCompleteTitle: "Квест выполнен",
    toastCompleteMsg: "Время истекло. Откройте Discord, чтобы забрать награду.",
    toastStoppedTitle: "Остановлено",
    toastStoppedMsg: "Окно игры было закрыто пользователем",
    toastInputErrTitle: "Ошибка ввода",
    toastInputErrMsg: "Укажите имя исполняемого файла (.exe)",
    toastSavedTitle: "Сохранено",
    toastSavedMsg: (exe) => `Процесс ${exe} добавлен в список`,
    toastLaunchErrTitle: "Ошибка запуска",
    toastConnErrTitle: "Ошибка связи",
    toastConnErrMsg: "Не удалось связаться с сервисом: ",
    soundTitle: "Звук по завершению",
  }
};

let currentLang = localStorage.getItem("dqh_lang_v2") || "en";

function t(key, arg) {
  const dict = i18n[currentLang] || i18n.en;
  const val = dict[key];
  if (typeof val === "function") return val(arg);
  return val !== undefined ? val : key;
}

// State
let allGames = [];
let selectedGame = null;
let isRunning = false;
let statusInterval = null;
let pollCount = 0;
let isSoundEnabled = localStorage.getItem("dqh_sound") !== "false";
let detectedQuests = [];
let matchedQuestsByGameId = {};
let selectedDurationSeconds = 15 * 60 + 15;

const CIRCUMFERENCE = 2 * Math.PI * 82; // 515.22

// Elements
const brandLogo = document.getElementById("brandLogo");
const openDiscordHeaderBtn = document.getElementById("openDiscordHeaderBtn");
const openDiscordBtnText = document.getElementById("openDiscordBtnText");
const cacheTimeText = document.getElementById("cacheTimeText");
const cacheDot = document.getElementById("cacheDot");

const questProcessesHeading = document.getElementById("questProcessesHeading");
const customProcessHeading = document.getElementById("customProcessHeading");
const startBtnText = document.getElementById("startBtnText");
const stopBtnText = document.getElementById("stopBtnText");

const gamesListEl = document.getElementById("gamesList");
const gamesCountEl = document.getElementById("gamesCount");
const searchInput = document.getElementById("searchInput");
const selectedGameTitle = document.getElementById("selectedGameTitle");
const selectedGameExe = document.getElementById("selectedGameExe");
const selectedGameQuestTag = document.getElementById("selectedGameQuestTag");
const selectedGameHint = document.getElementById("selectedGameHint");
const statusBadge = document.getElementById("statusBadge");
const statusText = document.getElementById("statusText");
const ringFill = document.getElementById("ringFill");
const timerDigits = document.getElementById("timerDigits");
const timerCaption = document.getElementById("timerCaption");
const startBtn = document.getElementById("startBtn");
const stopBtn = document.getElementById("stopBtn");

const customExeInput = document.getElementById("customExeInput");
const customTitleInput = document.getElementById("customTitleInput");
const customSelectBtn = document.getElementById("customSelectBtn");

const langToggleBtn = document.getElementById("langToggleBtn");
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

  applyLanguage(currentLang);
  updateSoundUI();
  await fetchGames();
  await fetchQuests(true);
  setupEventListeners();
  checkInitialStatus();
}

function applyLanguage(lang) {
  currentLang = lang;
  localStorage.setItem("dqh_lang_v2", lang);
  document.documentElement.lang = lang;
  const dict = i18n[lang] || i18n.en;

  if (langToggleBtn) langToggleBtn.textContent = dict.langCode;
  if (questProcessesHeading) questProcessesHeading.textContent = dict.questProcesses;
  if (searchInput) searchInput.placeholder = dict.searchPlaceholder;
  if (customProcessHeading) customProcessHeading.textContent = dict.customProcess;
  if (customSelectBtn) customSelectBtn.textContent = dict.selectBtn;
  if (openDiscordBtnText) openDiscordBtnText.textContent = dict.openDiscord;
  if (soundToggleBtn) soundToggleBtn.title = dict.soundTitle;

  if (isRunning) {
    statusText.textContent = dict.statusRunning;
    timerCaption.textContent = dict.captionRunning(selectedGame ? selectedGame.exe : "");
    if (stopBtnText) stopBtnText.textContent = dict.stopBtn;
  } else {
    statusText.textContent = dict.statusIdle;
    timerCaption.textContent = dict.captionIdle;
    if (startBtnText) {
      if (selectedDurationSeconds && selectedDurationSeconds < 900) {
        startBtnText.textContent = t("startBtnRem", formatMinSec(selectedDurationSeconds));
      } else {
        startBtnText.textContent = dict.startBtn;
      }
    }
  }

  if (allGames.length > 0) {
    renderGames(getFilteredGames());
    updateActiveGameQuestBadge();
  }
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

// Window size clamp
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

function formatMinSec(sec) {
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

function getQuestScore(game) {
  const q = matchedQuestsByGameId[game.id];
  if (!q) return 0;
  if (q.enrolled && !q.completed) return 3; // In progress first
  if (q.completed && !q.claimed) return 2;  // Ready to claim next
  if (q.claimed) return 1;                  // Already claimed
  return 0;
}

function sortGamesWithQuests() {
  allGames.sort((a, b) => {
    const scoreDiff = getQuestScore(b) - getQuestScore(a);
    if (scoreDiff !== 0) return scoreDiff;
    return a.title.localeCompare(b.title);
  });
}

function getFilteredGames() {
  const q = searchInput ? searchInput.value.toLowerCase().trim() : "";
  if (!q) return allGames;
  return allGames.filter(
    (g) =>
      g.title.toLowerCase().includes(q) ||
      g.exe.toLowerCase().includes(q)
  );
}

async function fetchGames() {
  try {
    const res = await fetch("/api/games");
    const serverGames = await res.json();
    const savedCustom = loadCustomGames();
    allGames = [...serverGames, ...savedCustom];
    if (gamesCountEl) gamesCountEl.textContent = allGames.length;
    sortGamesWithQuests();
    renderGames(getFilteredGames());
    if (allGames.length > 0 && !selectedGame) {
      selectGame(allGames[0]);
    }
  } catch (err) {
    console.error("Failed to load games:", err);
  }
}

async function fetchQuests(silent = false) {
  try {
    const res = await fetch("/api/quests");
    const data = await res.json();
    detectedQuests = data.quests || [];
    matchedQuestsByGameId = {};

    let needsGamesReload = false;
    for (const q of detectedQuests) {
      if (q.matchedGameId) {
        const existing = matchedQuestsByGameId[q.matchedGameId];
        if (!existing || (q.enrolled && !q.completed) || (existing.claimed && !q.claimed)) {
          matchedQuestsByGameId[q.matchedGameId] = q;
        }
        if (!allGames.some((g) => g.id === q.matchedGameId)) {
          needsGamesReload = true;
        }
      }
    }

    if (needsGamesReload) {
      await fetchGames();
      return;
    }

    sortGamesWithQuests();
    renderGames(getFilteredGames());
    updateActiveGameQuestBadge();

    if (!selectedGame && allGames.length > 0) {
      selectGame(allGames[0]);
    }

    // Update footer cache timestamp
    if (data.foundCache && data.modUnix) {
      const d = new Date(data.modUnix * 1000);
      const tStr = d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
      if (cacheTimeText) cacheTimeText.textContent = t("cacheStatusSynced", tStr);
      if (cacheDot) cacheDot.style.background = "var(--green)";
    } else {
      if (cacheTimeText) cacheTimeText.textContent = t("cacheStatusNotFound");
      if (cacheDot) cacheDot.style.background = "var(--text-muted)";
    }
  } catch (err) {
    console.error("Failed to sync quests:", err);
  }
}

function renderGames(games) {
  if (!gamesListEl) return;
  gamesListEl.innerHTML = "";
  games.forEach((game) => {
    const isSelected = selectedGame && selectedGame.id === game.id;
    const card = document.createElement("div");
    card.className = `game-card ${isSelected ? "active" : ""}`;

    const quest = matchedQuestsByGameId[game.id];
    let badgeHTML = "";

    if (quest) {
      if (quest.completed && !quest.claimed) {
        badgeHTML = `<span class="quest-tag ready">${t("questBadgeReady")}</span>`;
      } else if (quest.enrolled && !quest.completed) {
        const curM = formatMinSec(quest.progressSeconds);
        const totM = formatMinSec(quest.targetSeconds);
        badgeHTML = `<span class="quest-tag in-progress">${curM} / ${totM}</span>`;
      } else if (quest.claimed) {
        badgeHTML = `<span class="quest-tag claimed">${t("questBadgeClaimed")}</span>`;
      }
    }

    card.innerHTML = `
      <div class="game-info">
        <div class="game-title">${game.title}</div>
        <div class="game-exe">${game.exe}</div>
      </div>
      ${badgeHTML}
    `;
    card.onclick = () => selectGame(game);
    gamesListEl.appendChild(card);
  });
}

function updateActiveGameQuestBadge() {
  if (!selectedGame) return;
  const quest = matchedQuestsByGameId[selectedGame.id];

  if (quest && quest.enrolled && !quest.completed) {
    if (selectedGameQuestTag) {
      selectedGameQuestTag.className = "quest-status-inline in-progress";
      selectedGameQuestTag.textContent = `${formatMinSec(quest.progressSeconds)} / ${formatMinSec(quest.targetSeconds)}`;
    }
    if (selectedGameHint && !isRunning) {
      const bufferSec = 15;
      const recommended = Math.max(30, quest.remainingSeconds + bufferSec);
      selectedGameHint.textContent = `• ${t("smartPreFillHint", formatMinSec(recommended))}`;
    }
  } else if (quest && quest.completed && !quest.claimed) {
    if (selectedGameQuestTag) {
      selectedGameQuestTag.className = "quest-status-inline ready";
      selectedGameQuestTag.textContent = t("questBadgeReady");
    }
    if (selectedGameHint && !isRunning) {
      selectedGameHint.textContent = `• ${t("readyToClaimHint")}`;
    }
  } else if (quest && quest.claimed) {
    if (selectedGameQuestTag) {
      selectedGameQuestTag.className = "quest-status-inline claimed";
      selectedGameQuestTag.textContent = t("questBadgeClaimed");
    }
    if (selectedGameHint && !isRunning) {
      selectedGameHint.textContent = "";
    }
  } else {
    if (selectedGameQuestTag) {
      selectedGameQuestTag.className = "quest-status-inline hidden";
    }
    if (selectedGameHint && !isRunning) {
      selectedGameHint.textContent = "";
    }
  }
}

function selectGame(game) {
  if (isRunning) return;
  selectedGame = game;
  selectedGameTitle.textContent = game.title;
  selectedGameExe.textContent = game.exe;

  document.querySelectorAll(".game-card").forEach((c) => c.classList.remove("active"));
  const cards = Array.from(document.querySelectorAll(".game-card"));
  const activeCard = cards.find(
    (c) => c.querySelector(".game-title") && c.querySelector(".game-title").textContent === game.title
  );
  if (activeCard) activeCard.classList.add("active");

  const quest = matchedQuestsByGameId[game.id];
  if (quest && quest.enrolled && !quest.completed) {
    const bufferSec = 15;
    const recommended = Math.max(30, quest.remainingSeconds + bufferSec);
    selectedDurationSeconds = recommended;
    updateTimerUI(selectedDurationSeconds, selectedDurationSeconds);
    if (startBtnText) {
      startBtnText.textContent = t("startBtnRem", formatMinSec(selectedDurationSeconds));
    }
  } else {
    selectedDurationSeconds = 15 * 60 + 15;
    resetTimerUI();
    if (startBtnText) startBtnText.textContent = t("startBtn");
  }

  updateActiveGameQuestBadge();
}

function setupEventListeners() {
  // Lang Toggle
  if (langToggleBtn) {
    langToggleBtn.addEventListener("click", () => {
      const next = currentLang === "en" ? "ru" : "en";
      applyLanguage(next);
    });
  }

  // Search
  searchInput.addEventListener("input", () => {
    renderGames(getFilteredGames());
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
    const defaultTitle = currentLang === "ru" ? "Свой процесс" : "Custom Process";
    const title = customTitleInput.value.trim() || defaultTitle;
    if (!exe) {
      showToast(t("toastInputErrTitle"), t("toastInputErrMsg"));
      return;
    }
    const cleanExe = exe.toLowerCase().endsWith(".exe") ? exe : `${exe}.exe`;
    const custom = {
      id: "custom_" + Date.now(),
      title: title,
      exe: cleanExe,
      windowTitle: title,
    };
    
    const saved = loadCustomGames();
    saved.push(custom);
    saveCustomGames(saved);

    allGames.push(custom);
    if (gamesCountEl) gamesCountEl.textContent = allGames.length;
    renderGames(getFilteredGames());
    selectGame(custom);
    customExeInput.value = "";
    customTitleInput.value = "";
    showToast(t("toastSavedTitle"), t("toastSavedMsg", cleanExe));
  });

  // Hotkey '/' to focus search
  window.addEventListener("keydown", (e) => {
    if (e.key === "/" && document.activeElement !== searchInput && document.activeElement !== customExeInput && document.activeElement !== customTitleInput) {
      e.preventDefault();
      searchInput.focus();
    }
  });

  // Buttons
  startBtn.addEventListener("click", startGame);
  stopBtn.addEventListener("click", stopGame);

  // Auto-sync whenever window gains focus (even when running)
  window.addEventListener("focus", () => {
    fetchQuests(true);
  });

  // Toast close
  toastCloseBtn.addEventListener("click", () => {
    toastEl.classList.add("hidden");
  });

  // Open Discord client directly to Quests tab
  const openDiscord = async (e) => {
    if (e && e.preventDefault) e.preventDefault();
    try {
      await fetch("/api/open-discord", { method: "POST" });
      setTimeout(() => fetchQuests(true), 2000);
    } catch (err) {
      console.error("Open discord error:", err);
    }
  };

  if (openDiscordHeaderBtn) openDiscordHeaderBtn.addEventListener("click", openDiscord);
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
        steamAppId: selectedGame.steamAppId ? String(selectedGame.steamAppId) : "",
        discordAppId: selectedGame.discordAppId ? String(selectedGame.discordAppId) : "",
        durationSeconds: selectedDurationSeconds || 915
      })
    });

    if (res.ok) {
      setRunningState(true);
      startPolling();
    } else {
      const errText = await res.text();
      showToast(t("toastLaunchErrTitle"), errText);
    }
  } catch (err) {
    showToast(t("toastConnErrTitle"), t("toastConnErrMsg") + err);
  }
}

async function stopGame() {
  if (!isRunning) return;

  try {
    const res = await fetch("/api/stop", { method: "POST" });
    if (res.ok) {
      setRunningState(false);
      resetTimerUI();
    }
  } catch (err) {
    showToast(t("toastConnErrTitle"), t("toastConnErrMsg") + err);
  }
}

function setRunningState(running) {
  isRunning = running;
  startBtn.disabled = running;
  stopBtn.disabled = !running;

  if (running) {
    if (brandLogo) brandLogo.classList.add("is-running");
    statusBadge.className = "status-indicator active";
    statusText.textContent = t("statusRunning");
    timerCaption.textContent = t("captionRunning", selectedGame ? selectedGame.exe : "");
    ringFill.classList.add("running");
    timerDigits.classList.add("running");
  } else {
    if (brandLogo) brandLogo.classList.remove("is-running");
    statusBadge.className = "status-indicator";
    statusText.textContent = t("statusIdle");
    timerCaption.textContent = t("captionIdle");
    ringFill.classList.remove("running");
    timerDigits.classList.remove("running");
    if (statusInterval) {
      clearInterval(statusInterval);
      statusInterval = null;
    }
  }
}

function startPolling() {
  if (statusInterval) clearInterval(statusInterval);
  pollCount = 0;
  statusInterval = setInterval(pollStatus, 1000);
}

async function pollStatus() {
  try {
    pollCount++;
    // Auto-refresh quest cache detection every 5 seconds while running
    if (pollCount % 5 === 0) {
      fetchQuests(true);
    }

    const res = await fetch("/api/status");
    const data = await res.json();

    if (!data.running && isRunning) {
      setRunningState(false);
      if (data.remaining <= 2) {
        onQuestCompleted();
      } else {
        resetTimerUI();
        showToast(t("toastStoppedTitle"), t("toastStoppedMsg"));
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
  const offset = CIRCUMFERENCE * (1 - fraction);
  ringFill.style.strokeDashoffset = offset;
}

function resetTimerUI() {
  timerDigits.textContent = "15:00";
  ringFill.style.strokeDashoffset = CIRCUMFERENCE;
}

function onQuestCompleted() {
  if (brandLogo) brandLogo.classList.remove("is-running");
  timerDigits.textContent = "00:00";
  ringFill.style.strokeDashoffset = 0;
  statusBadge.className = "status-indicator completed";
  statusText.textContent = t("statusCompleted");
  timerCaption.textContent = t("captionCompleted");

  if (isSoundEnabled) {
    playChime();
  }

  showToast(t("toastCompleteTitle"), t("toastCompleteMsg"));

  // Trigger auto refresh of Discord cache after completion
  setTimeout(() => fetchQuests(true), 2500);
}

function showToast(title, message) {
  toastTitle.textContent = title;
  toastMsg.textContent = message;
  toastEl.classList.remove("hidden");
  setTimeout(() => {
    toastEl.classList.add("hidden");
  }, 5000);
}

// Chime synthesized audio
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
