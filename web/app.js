// Localization Dictionaries
const i18n = {
  ru: {
    langCode: "RU",
    appSubtitle: "Эмуляция игровых процессов для наград Discord",
    questProcesses: "Процессы квестов",
    searchPlaceholder: "Поиск (Roblox, Call of Duty, RE7)...",
    customProcess: "Пользовательский процесс",
    customTitlePlaceholder: "Название (опционально)",
    selectBtn: "Выбрать",
    selectedGameLabel: "ТЕКУЩИЙ ВЫБОР",
    statusIdle: "В ожидании",
    statusRunning: "Эмуляция активна",
    statusCompleted: "Выполнено",
    timerSubTimer: "ТАЙМЕР",
    timerSubRemaining: "ОСТАЛОСЬ",
    timerSubCompleted: "РЕЗУЛЬТАТ",
    captionIdle: "Выберите процесс и нажмите «Запустить эмуляцию»",
    captionRunning: (exe) => `Discord видит процесс: ${exe}`,
    captionCompleted: "15 минут истекли. Награда готова к получению в Discord!",
    startBtn: "Запустить эмуляцию (15 мин)",
    stopBtn: "Остановить",
    hintText: "<strong>Подсказка:</strong> Запустите эмуляцию. Окно игры появится свёрнутым в панели задач. Через 15 минут прозвучит сигнал о награде.",
    openDiscord: "Открыть вкладку квестов",
    syncBtnTitle: "Синхронизировать квесты Discord",
    questBadgeReady: "Готов к получению",
    questBadgeClaimed: "Получен",
    questRemShort: (rem) => `осталось ${rem}`,
    bannerTagActive: "В процессе",
    bannerTagReady: "Готов к получению",
    bannerDescActive: (rem) => `Таймер настроен на ${rem} для гарантированного завершения.`,
    bannerDescReady: "Квест выполнен на 100%. Откройте вкладку квестов в Discord, чтобы забрать награду.",
    startBtnRem: (rem) => `Запустить квест (${rem})`,
    toastSyncTitle: "Квесты Discord",
    toastSyncFound: (n) => `Обнаружено активных квестов: ${n}`,
    toastSyncNotFound: "Кэш квестов Discord не найден",
    toastSyncUnchanged: "Кэш Discord еще не обновился. Откройте вкладку «Квесты» в Discord или нажмите Ctrl+R в Discord.",
    toastCompleteTitle: "Квест завершен!",
    toastCompleteMsg: "15 минут успешно завершены. Откройте Discord для получения награды.",
    toastStoppedTitle: "Эмуляция остановлена",
    toastStoppedMsg: "Окно игры было закрыто пользователем",
    toastInputErrTitle: "Ошибка ввода",
    toastInputErrMsg: "Укажите имя исполняемого файла процесса (.exe)",
    toastSavedTitle: "Процесс сохранён",
    toastSavedMsg: (exe) => `Процесс ${exe} добавлен в ваш список`,
    toastLaunchErrTitle: "Ошибка запуска",
    toastConnErrTitle: "Ошибка связи",
    toastConnErrMsg: "Не удалось связаться с локальным сервисом: ",
    soundTitle: "Звук по завершению",
  },
  en: {
    langCode: "EN",
    appSubtitle: "Game process emulator for Discord Quest rewards",
    questProcesses: "Quest Processes",
    searchPlaceholder: "Search (Roblox, Call of Duty, RE7)...",
    customProcess: "Custom Process",
    customTitlePlaceholder: "Title (optional)",
    selectBtn: "Select",
    selectedGameLabel: "SELECTED GAME",
    statusIdle: "Idle",
    statusRunning: "Emulating",
    statusCompleted: "Completed",
    timerSubTimer: "TIMER",
    timerSubRemaining: "REMAINING",
    timerSubCompleted: "RESULT",
    captionIdle: "Select a process and click \"Start Quest\"",
    captionRunning: (exe) => `Discord sees process: ${exe}`,
    captionCompleted: "15 minutes elapsed. Reward is ready to claim in Discord!",
    startBtn: "Start Quest (15 min)",
    stopBtn: "Stop",
    hintText: "<strong>Tip:</strong> Start emulation. The game window will launch minimized in the taskbar. After 15 minutes, a chime will notify you when rewards are ready.",
    openDiscord: "Open Discord Quests",
    syncBtnTitle: "Sync Discord Quests",
    questBadgeReady: "Ready to claim",
    questBadgeClaimed: "Claimed",
    questRemShort: (rem) => `${rem} left`,
    bannerTagActive: "In Progress",
    bannerTagReady: "Ready to claim",
    bannerDescActive: (rem) => `Timer configured to ${rem} to guarantee completion.`,
    bannerDescReady: "Quest completed 100%. Open Discord to claim your reward.",
    startBtnRem: (rem) => `Start Quest (${rem})`,
    toastSyncTitle: "Discord Quests",
    toastSyncFound: (n) => `Detected ${n} active quest(s)`,
    toastSyncNotFound: "Discord quest cache not found",
    toastSyncUnchanged: "Discord cache has not updated yet. Open the Quests tab in Discord or press Ctrl+R to refresh.",
    toastCompleteTitle: "Quest Completed!",
    toastCompleteMsg: "15 minutes elapsed. Open Discord to claim your reward.",
    toastStoppedTitle: "Emulation Stopped",
    toastStoppedMsg: "Game simulation window was closed by the user",
    toastInputErrTitle: "Input Error",
    toastInputErrMsg: "Please specify an executable process name (.exe)",
    toastSavedTitle: "Process Saved",
    toastSavedMsg: (exe) => `Process ${exe} saved to your list`,
    toastLaunchErrTitle: "Launch Error",
    toastConnErrTitle: "Connection Error",
    toastConnErrMsg: "Failed to communicate with local service: ",
    soundTitle: "Completion sound",
  }
};

let currentLang = localStorage.getItem("dqh_lang") || "en";

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
let isSoundEnabled = localStorage.getItem("dqh_sound") !== "false";
let detectedQuests = [];
let matchedQuestsByGameId = {};
let selectedDurationSeconds = 15 * 60 + 15;
let lastCacheModUnix = 0;

// SVG Icons (replacing cheap text emojis)
const ICON_TARGET_SVG = `<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="22" y1="12" x2="18" y2="12"/><line x1="6" y1="12" x2="2" y2="12"/><line x1="12" y1="6" x2="12" y2="2"/><line x1="12" y1="22" x2="12" y2="18"/></svg>`;
const ICON_CHECK_SVG = `<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>`;
const ICON_REWARD_SVG = `<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M20 12V8H6a2 2 0 0 1-2-2c0-1.1.9-2 2-2h12v4"/><path d="M4 6v12a2 2 0 0 0 2 2h14v-4"/><circle cx="18" cy="12" r="2"/></svg>`;

const BANNER_TARGET_SVG = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="22" y1="12" x2="18" y2="12"/><line x1="6" y1="12" x2="2" y2="12"/><line x1="12" y1="6" x2="12" y2="2"/><line x1="12" y1="22" x2="12" y2="18"/></svg>`;
const BANNER_CHECK_SVG = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>`;

const CIRCUMFERENCE = 2 * Math.PI * 82; // 515.22

// Elements
const brandLogo = document.getElementById("brandLogo");
const syncQuestsBtn = document.getElementById("syncQuestsBtn");
const syncIcon = document.getElementById("syncIcon");
const questSmartBanner = document.getElementById("questSmartBanner");
const bannerBadgeIcon = document.getElementById("bannerBadgeIcon");
const bannerTitle = document.getElementById("bannerTitle");
const bannerTag = document.getElementById("bannerTag");
const bannerDesc = document.getElementById("bannerDesc");

const appSubtitle = document.getElementById("appSubtitle");
const questProcessesHeading = document.getElementById("questProcessesHeading");
const customProcessHeading = document.getElementById("customProcessHeading");
const selectedGameLabel = document.getElementById("selectedGameLabel");
const startBtnText = document.getElementById("startBtnText");
const stopBtnText = document.getElementById("stopBtnText");
const infoTipText = document.getElementById("infoTipText");
const openDiscordBtnText = document.getElementById("openDiscordBtnText");

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
  localStorage.setItem("dqh_lang", lang);
  document.documentElement.lang = lang;
  const dict = i18n[lang] || i18n.en;

  if (langToggleBtn) langToggleBtn.textContent = dict.langCode;
  if (appSubtitle) appSubtitle.textContent = dict.appSubtitle;
  if (questProcessesHeading) questProcessesHeading.textContent = dict.questProcesses;
  if (searchInput) searchInput.placeholder = dict.searchPlaceholder;
  if (customProcessHeading) customProcessHeading.textContent = dict.customProcess;
  if (customTitleInput) customTitleInput.placeholder = dict.customTitlePlaceholder;
  if (customSelectBtn) customSelectBtn.textContent = dict.selectBtn;
  if (selectedGameLabel) selectedGameLabel.textContent = dict.selectedGameLabel;
  if (infoTipText) infoTipText.innerHTML = dict.hintText;
  if (openDiscordBtnText) openDiscordBtnText.textContent = dict.openDiscord;
  if (soundToggleBtn) soundToggleBtn.title = dict.soundTitle;
  if (syncQuestsBtn) syncQuestsBtn.title = dict.syncBtnTitle;

  if (isRunning) {
    statusText.textContent = dict.statusRunning;
    timerSub.textContent = dict.timerSubRemaining;
    timerCaption.textContent = dict.captionRunning(selectedGame ? selectedGame.exe : "");
    if (stopBtnText) stopBtnText.textContent = dict.stopBtn;
  } else {
    statusText.textContent = dict.statusIdle;
    timerSub.textContent = dict.timerSubTimer;
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
    renderGames(allGames);
    if (selectedGame) selectGame(selectedGame);
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

async function fetchGames() {
  try {
    const res = await fetch("/api/games");
    const serverGames = await res.json();
    const savedCustom = loadCustomGames();
    allGames = [...serverGames, ...savedCustom];
    gamesCountEl.textContent = allGames.length;
    sortGamesWithQuests();
    renderGames(allGames);
    if (allGames.length > 0 && !selectedGame) {
      selectGame(allGames[0]);
    }
  } catch (err) {
    console.error("Failed to load games:", err);
  }
}

async function fetchQuests(silent = false) {
  if (syncIcon) syncIcon.classList.add("spin");
  try {
    const res = await fetch("/api/quests");
    const data = await res.json();
    detectedQuests = data.quests || [];
    matchedQuestsByGameId = {};

    for (const q of detectedQuests) {
      if (q.matchedGameId) {
        matchedQuestsByGameId[q.matchedGameId] = q;
      }
    }

    sortGamesWithQuests();
    renderGames(allGames);

    if (selectedGame) {
      selectGame(selectedGame);
    } else if (allGames.length > 0) {
      selectGame(allGames[0]);
    }

    if (!silent) {
      if (!data.foundCache) {
        showToast(t("toastSyncTitle"), t("toastSyncNotFound"));
      } else if (lastCacheModUnix > 0 && data.modUnix === lastCacheModUnix) {
        showToast(t("toastSyncTitle"), t("toastSyncUnchanged"));
      } else {
        showToast(t("toastSyncTitle"), t("toastSyncFound", detectedQuests.length));
      }
    }
    lastCacheModUnix = data.modUnix || 0;
  } catch (err) {
    console.error("Failed to sync quests:", err);
    if (!silent) {
      showToast(t("toastSyncTitle"), t("toastConnErrMsg") + err);
    }
  } finally {
    if (syncIcon) {
      setTimeout(() => syncIcon.classList.remove("spin"), 500);
    }
  }
}

function renderGames(games) {
  gamesListEl.innerHTML = "";
  games.forEach((game) => {
    const card = document.createElement("div");
    card.className = `game-card ${selectedGame && selectedGame.id === game.id ? "active" : ""}`;

    const quest = matchedQuestsByGameId[game.id];
    let badgeHTML = "";
    let footerExtra = "";

    if (quest) {
      if (quest.completed && !quest.claimed) {
        badgeHTML = `<span class="quest-pill ready" title="${t("questBadgeReady")}"><span class="quest-pulse-dot"></span>${ICON_CHECK_SVG}<span>${t("questBadgeReady")}</span></span>`;
      } else if (quest.enrolled && !quest.completed) {
        const curM = formatMinSec(quest.progressSeconds);
        const totM = formatMinSec(quest.targetSeconds);
        badgeHTML = `<span class="quest-pill in-progress" title="${t("bannerTagActive")}"><span class="quest-pulse-dot"></span>${ICON_TARGET_SVG}<span>${curM} / ${totM}</span></span>`;
        footerExtra = `<span class="quest-rem-pill">${t("questRemShort", formatMinSec(quest.remainingSeconds))}</span>`;
      } else if (quest.claimed) {
        badgeHTML = `<span class="quest-pill claimed" title="${t("questBadgeClaimed")}">${ICON_REWARD_SVG}<span>${t("questBadgeClaimed")}</span></span>`;
      }
    }

    card.innerHTML = `
      <div class="game-card-left">
        <div class="game-header-row">
          <div class="game-title">${game.title}</div>
          ${badgeHTML}
        </div>
        <div class="game-footer-row">
          <div class="game-exe-pill">${game.exe}</div>
          ${footerExtra}
        </div>
      </div>
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
    (c) => c.querySelector(".game-title") && c.querySelector(".game-title").textContent === game.title
  );
  if (activeCard) activeCard.classList.add("active");

  const quest = matchedQuestsByGameId[game.id];
  if (quest && quest.enrolled && !quest.completed) {
    const bufferSec = 15;
    const recommended = Math.max(30, quest.remainingSeconds + bufferSec);
    selectedDurationSeconds = recommended;

    if (questSmartBanner) {
      questSmartBanner.className = "quest-smart-banner";
      bannerBadgeIcon.innerHTML = BANNER_TARGET_SVG;
      bannerTitle.textContent = quest.questName || game.title;
      bannerTag.textContent = t("bannerTagActive");
      bannerDesc.textContent = t("bannerDescActive", formatMinSec(recommended));
    }

    updateTimerUI(selectedDurationSeconds, selectedDurationSeconds);
    if (startBtnText) {
      startBtnText.textContent = t("startBtnRem", formatMinSec(selectedDurationSeconds));
    }
  } else if (quest && quest.completed && !quest.claimed) {
    selectedDurationSeconds = 15 * 60 + 15;
    if (questSmartBanner) {
      questSmartBanner.className = "quest-smart-banner ready";
      bannerBadgeIcon.innerHTML = BANNER_CHECK_SVG;
      bannerTitle.textContent = quest.questName || game.title;
      bannerTag.textContent = t("bannerTagReady");
      bannerDesc.textContent = t("bannerDescReady");
    }
    resetTimerUI();
    if (startBtnText) startBtnText.textContent = t("startBtn");
  } else {
    selectedDurationSeconds = 15 * 60 + 15;
    if (questSmartBanner) {
      questSmartBanner.className = "quest-smart-banner hidden";
    }
    resetTimerUI();
    if (startBtnText) startBtnText.textContent = t("startBtn");
  }
}

function setupEventListeners() {
  // Lang Toggle
  if (langToggleBtn) {
    langToggleBtn.addEventListener("click", () => {
      const next = currentLang === "ru" ? "en" : "ru";
      applyLanguage(next);
    });
  }

  // Search
  searchInput.addEventListener("input", (e) => {
    const q = e.target.value.toLowerCase().trim();
    const filtered = allGames.filter(
      (g) =>
        g.title.toLowerCase().includes(q) ||
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
    gamesCountEl.textContent = allGames.length;
    renderGames(allGames);
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

  // Sync Quests
  if (syncQuestsBtn) {
    syncQuestsBtn.addEventListener("click", () => fetchQuests(false));
  }

  // Auto-sync when window gains focus
  window.addEventListener("focus", () => {
    if (!isRunning) fetchQuests(true);
  });

  // Toast close
  toastCloseBtn.addEventListener("click", () => {
    toastEl.classList.add("hidden");
  });

  // Open Discord client
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
        discordAppId: selectedGame.discordAppId || "",
        durationSeconds: selectedDurationSeconds || 915,
      }),
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
    if (brandLogo) {
      brandLogo.classList.add("is-running");
      brandLogo.classList.remove("is-completed");
    }
    statusBadge.className = "status-indicator active";
    statusText.textContent = t("statusRunning");
    timerCaption.textContent = t("captionRunning", selectedGame ? selectedGame.exe : "");
    ringFill.classList.add("running");
    timerDigits.classList.add("running");
    timerSub.textContent = t("timerSubRemaining");
  } else {
    if (brandLogo) {
      brandLogo.classList.remove("is-running");
    }
    statusBadge.className = "status-indicator";
    statusText.textContent = t("statusIdle");
    timerCaption.textContent = t("captionIdle");
    ringFill.classList.remove("running");
    timerDigits.classList.remove("running");
    timerSub.textContent = t("timerSubTimer");
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
  if (brandLogo) {
    brandLogo.classList.remove("is-running");
  }
  timerDigits.textContent = "00:00";
  timerPercent.textContent = "100%";
  ringFill.style.strokeDashoffset = 0;
  statusBadge.className = "status-indicator completed";
  statusText.textContent = t("statusCompleted");
  timerCaption.textContent = t("captionCompleted");
  timerSub.textContent = t("timerSubCompleted");

  if (isSoundEnabled) {
    playChime();
  }

  showToast(t("toastCompleteTitle"), t("toastCompleteMsg"));

  // Trigger auto refresh of Discord cache after completion
  setTimeout(() => fetchQuests(true), 3500);
}

function showToast(title, message) {
  toastTitle.textContent = title;
  toastMsg.textContent = message;
  toastEl.classList.remove("hidden");
  setTimeout(() => {
    toastEl.classList.add("hidden");
  }, 7000);
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
