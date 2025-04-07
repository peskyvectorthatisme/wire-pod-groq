const intentsJson = JSON.parse(
  '["intent_greeting_hello", "intent_names_ask", "intent_imperative_eyecolor", "intent_character_age", "intent_explore_start", "intent_system_charger", "intent_system_sleep", "intent_greeting_goodmorning", "intent_greeting_goodnight", "intent_greeting_goodbye", "intent_seasonal_happynewyear", "intent_seasonal_happyholidays", "intent_amazon_signin", "intent_imperative_forward", "intent_imperative_turnaround", "intent_imperative_turnleft", "intent_imperative_turnright", "intent_play_rollcube", "intent_play_popawheelie", "intent_play_fistbump", "intent_play_blackjack", "intent_imperative_affirmative", "intent_imperative_negative", "intent_photo_take_extend", "intent_imperative_praise", "intent_imperative_abuse", "intent_weather_extend", "intent_imperative_apologize", "intent_imperative_backup", "intent_imperative_volumedown", "intent_imperative_volumeup", "intent_imperative_lookatme", "intent_imperative_volumelevel_extend", "intent_imperative_shutup", "intent_names_username_extend", "intent_imperative_come", "intent_imperative_love", "intent_knowledge_promptquestion", "intent_clock_checktimer", "intent_global_stop_extend", "intent_clock_settimer_extend", "intent_clock_time", "intent_imperative_quiet", "intent_imperative_dance", "intent_play_pickupcube", "intent_imperative_fetchcube", "intent_imperative_findcube", "intent_play_anytrick", "intent_message_recordmessage_extend", "intent_message_playmessage_extend", "intent_blackjack_hit", "intent_blackjack_stand", "intent_play_keepaway"]'
);

var GetLog = false;

const getE = (element) => document.getElementById(element);

function updateIntentSelection(element) {
  fetch("/api/get_custom_intents_json")
    .then((response) => response.json())
    .then((listResponse) => {
      const container = getE(element);
      container.innerHTML = "";
      if (listResponse && listResponse.length > 0) {
        const select = document.createElement("select");
        select.name = `${element}intents`;
        select.id = `${element}intents`;
        listResponse.forEach((intent) => {
          if (!intent.issystem) {
            const option = document.createElement("option");
            option.value = intent.name;
            option.text = intent.name;
            select.appendChild(option);
          }
        });
        const label = document.createElement("label");
        label.innerHTML = "Choose the intent: ";
        label.htmlFor = `${element}intents`;
        container.appendChild(label).appendChild(select);

        select.addEventListener("change", hideEditIntents);
      } else {
        const error = document.createElement("p");
        error.innerHTML = "No intents found, you must add one first";
        container.appendChild(error);
      }
    }).catch(() => {
      // Do nothing
    });
}

function checkInited() {
  fetch("/api/is_api_v3").then((response) => {
    if (!response.ok) {
      alert(
        "This webroot does not match with the wire-pod binary. Some functionality will be broken. There was either an error during the last update, or you did not precisely follow the update guide. https://github.com/kercre123/wire-pod/wiki/Things-to-Know#updating-wire-pod"
      );
    }
  });

  fetch("/api/get_config")
    .then((response) => response.json())
    .then((config) => {
      if (!config.pastinitialsetup) {
        window.location.href = "/initial.html";
      }
    });
}

function createIntentSelect(element) {
  const select = document.createElement("select");
  select.name = `${element}intents`;
  select.id = `${element}intents`;
  intentsJson.forEach((intent) => {
    const option = document.createElement("option");
    option.value = intent;
    option.text = intent;
    select.appendChild(option);
  });
  const label = document.createElement("label");
  label.innerHTML = "Intent to send to robot after script executed:";
  label.htmlFor = `${element}intents`;
  getE(element).innerHTML = "";
  getE(element).appendChild(label).appendChild(select);
}

function editFormCreate() {
  const intentNumber = getE("editSelectintents").selectedIndex;

  fetch("/api/get_custom_intents_json")
    .then((response) => response.json())
    .then((intents) => {
      const intent = intents[intentNumber];
      if (intent) {
        const form = document.createElement("form");
        form.id = "editIntentForm";
        form.name = "editIntentForm";
        form.innerHTML = `
          <label for="name">Name:<br><input type="text" id="name" value="${intent.name}"></label><br>
          <label for="description">Description:<br><input type="text" id="description" value="${intent.description}"></label><br>
          <label for="utterances">Utterances:<br><input type="text" id="utterances" value="${intent.utterances.join(",")}"></label><br>
          <label for="intent">Intent:<br><select id="intent">${intentsJson
            .map(
              (name) =>
                `<option value="${name}" ${name === intent.intent ? "selected" : ""
                }>${name}</option>`
            )
            .join("")}</select></label><br>
          <label for="paramname">Param Name:<br><input type="text" id="paramname" value="${intent.params.paramname}"></label><br>
          <label for="paramvalue">Param Value:<br><input type="text" id="paramvalue" value="${intent.params.paramvalue}"></label><br>
          <label for="exec">Exec:<br><input type="text" id="exec" value="${intent.exec}"></label><br>
          <label for="execargs">Exec Args:<br><input type="text" id="execargs" value="${intent.execargs.join(",")}"></label><br>
          <label for="luascript">Lua code to run:</label><br><textarea id="luascript">${intent.luascript}</textarea>
          <button onclick="editIntent(${intentNumber})">Submit</button>
        `;
        //form.querySelector("#submit").onclick = () => editIntent(intentNumber);
        getE("editIntentForm").innerHTML = "";
        getE("editIntentForm").appendChild(form);
        showEditIntents();
      } else {
        displayError("editIntentForm", "No intents found, you must add one first");
      }
    }).catch((error) => {
      console.error(error);
      displayError("editIntentForm", "Error fetching intents");
    })
}

function editIntent(intentNumber) {
  const data = {
    number: intentNumber + 1,
    name: getE("name").value,
    description: getE("description").value,
    utterances: getE("utterances").value.split(","),
    intent: getE("intent").value,
    params: {
      paramname: getE("paramname").value,
      paramvalue: getE("paramvalue").value,
    },
    exec: getE("exec").value,
    execargs: getE("execargs").value.split(","),
    luascript: getE("luascript").value,
  };

  fetch("/api/edit_custom_intent", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      displayMessage("editIntentStatus", response);
      alert(response)
      updateIntentSelection("editSelect");
      updateIntentSelection("deleteSelect");
    });
}

function deleteSelectedIntent() {
  const intentNumber = getE("editSelectintents").selectedIndex + 1;

  fetch("/api/remove_custom_intent", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ number: intentNumber }),
  })
    .then((response) => response.text())
    .then((response) => {
      hideEditIntents();
      alert(response)
      updateIntentSelection("editSelect");
      updateIntentSelection("deleteSelect");
    });
}

function sendIntentAdd() {
  const form = getE("intentAddForm");
  const data = {
    name: form.elements["nameAdd"].value,
    description: form.elements["descriptionAdd"].value,
    utterances: form.elements["utterancesAdd"].value.split(","),
    intent: form.elements["intentAddSelectintents"].value,
    params: {
      paramname: form.elements["paramnameAdd"].value,
      paramvalue: form.elements["paramvalueAdd"].value,
    },
    exec: form.elements["execAdd"].value,
    execargs: form.elements["execAddArgs"].value.split(","),
    luascript: form.elements["luaAdd"].value,
  };
  if (!data.name || !data.description || !data.utterances) {
    displayMessage("addIntentStatus", "A required input is missing. You need a name, description, and utterances.");
    alert("A required input is missing. You need a name, description, and utterances.")
    return
  }

  displayMessage("addIntentStatus", "Adding...");

  fetch("/api/add_custom_intent", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      displayMessage("addIntentStatus", response);
      alert(response)
      updateIntentSelection("editSelect");
      updateIntentSelection("deleteSelect");
    });
}

function checkWeather() {
  getE("apiKeySpan").style.display = getE("weatherProvider").value ? "block" : "none";
}

function sendWeatherAPIKey() {
  const data = {
    provider: getE("weatherProvider").value,
    key: getE("apiKey").value,
  };

  displayMessage("addWeatherProviderAPIStatus", "Saving...");

  fetch("/api/set_weather_api", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      displayMessage("addWeatherProviderAPIStatus", response);
    });
}

function updateWeatherAPI() {
  fetch("/api/get_weather_api")
    .then((response) => response.json())
    .then((data) => {
      getE("weatherProvider").value = data.provider;
      getE("apiKey").value = data.key;
      checkWeather();
    });
}

function checkKG() {
  const provider = getE("kgProvider").value;
  const elements = [
    "houndifyInput",
    "togetherInput",
    "customAIInput",
    "intentGraphInput",
    "openAIInput",
    "saveChatInput",
    "llmCommandInput",
    "openAIVoiceForEnglishInput",
    "groqAIInput",
  ];

  elements.forEach((el) => (getE(el).style.display = "none"));

  if (provider) {
    if (provider === "houndify") {
      getE("houndifyInput").style.display = "block";
    } else if (provider === "openai") {
      getE("intentGraphInput").style.display = "block";
      getE("openAIInput").style.display = "block";
      getE("saveChatInput").style.display = "block";
      getE("llmCommandInput").style.display = "block";
      getE("openAIVoiceForEnglishInput").style.display = "block";
    } else if (provider === "together") {
      getE("intentGraphInput").style.display = "block";
      getE("togetherInput").style.display = "block";
      getE("saveChatInput").style.display = "block";
      getE("llmCommandInput").style.display = "block";
    } else if (provider === "custom") {
      getE("intentGraphInput").style.display = "block";
      getE("customAIInput").style.display = "block";
      getE("saveChatInput").style.display = "block";
      getE("llmCommandInput").style.display = "block";
    } else if (provider === "groq") {
      getE("intentGraphInput").style.display = "block";
      getE("groqAIInput").style.display = "block";
      getE("saveChatInput").style.display = "block";
      getE("llmCommandInput").style.display = "block";
    }
  }
}

function sendKGAPIKey() {
  const provider = getE("kgProvider").value;
  const data = {
    enable: true,
    provider,
    key: "",
    model: "",
    id: "",
    intentgraph: false,
    robotName: "",
    openai_prompt: "",
    openai_voice: "",
    openai_voice_with_english: false,
    save_chat: false,
    commands_enable: false,
    endpoint: "",
  };
  if (provider === "openai") {
    data.key = getE("openaiKey").value;
    data.openai_prompt = getE("openAIPrompt").value;
    data.intentgraph = getE("intentyes").checked
    data.save_chat = getE("saveChatYes").checked
    data.commands_enable = getE("commandYes").checked
    data.openai_voice = getE("openaiVoice").value
    data.openai_voice_with_english = getE("voiceEnglishYes").checked
  } else if (provider === "custom") {
    data.key = getE("customKey").value;
    data.model = getE("customModel").value;
    data.openai_prompt = getE("customAIPrompt").value;
    data.endpoint = getE("customAIEndpoint").value;
    data.intentgraph = getE("intentyes").checked
    data.save_chat = getE("saveChatYes").checked
    data.commands_enable = getE("commandYes").checked
  } else if (provider === "together") {
    data.key = getE("togetherKey").value;
    data.model = getE("togetherModel").value;
    data.openai_prompt = getE("togetherAIPrompt").value;
    data.intentgraph = getE("intentyes").checked;
    data.save_chat = getE("saveChatYes").checked
    data.commands_enable = getE("commandYes").checked
  } else if (provider === "groq") {
    data.key = getE("groqKey").value;
    data.model = getE("groqModel").value;
    data.openai_prompt = getE("groqAIPrompt").value;
    data.endpoint = getE("groqSpeechEndpoint").value;
    data.intentgraph = getE("intentyes").checked;
    data.save_chat = getE("saveChatYes").checked;
    data.commands_enable = getE("commandYes").checked;
  } else if (provider === "houndify") {
    data.key = getE("houndKey").value;
    data.id = getE("houndID").value;
  } else {
    data.enable = false;
  }

  fetch("/api/set_kg_api", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      displayMessage("addKGProviderAPIStatus", response);
      alert(response);
    });
}

function deleteSavedChats() {
  if (confirm("Are you sure? This will delete all saved chats.")) {
    fetch("/api/delete_chats")
      .then((response) => response.text())
      .then(() => {
        alert("Successfully deleted all saved chats.");
      });
  }
}

function updateKGAPI() {
  fetch("/api/get_kg_api")
    .then((response) => response.json())
    .then((data) => {
      getE("kgProvider").value = data.provider;
      if (data.provider === "openai") {
        getE("openaiKey").value = data.key;
        getE("openAIPrompt").value = data.openai_prompt;
        getE("openaiVoice").value = data.openai_voice;
        getE("commandYes").checked = data.commands_enable
        getE("intentyes").checked = data.intentgraph
        getE("saveChatYes").checked = data.save_chat
        getE("voiceEnglishYes").checked = data.openai_voice_with_english
      } else if (data.provider === "together") {
        getE("togetherKey").value = data.key;
        getE("togetherModel").value = data.model;
        getE("togetherAIPrompt").value = data.openai_prompt;
        getE("commandYes").checked = data.commands_enable
        getE("intentyes").checked = data.intentgraph
        getE("saveChatYes").checked = data.save_chat
      } else if (data.provider === "custom") {
        getE("customKey").value = data.key;
        getE("customModel").value = data.model;
        getE("customAIPrompt").value = data.openai_prompt;
        getE("customAIEndpoint").value = data.endpoint;
        getE("commandYes").checked = data.commands_enable
        getE("intentyes").checked = data.intentgraph
        getE("saveChatYes").checked = data.save_chat
      } else if (data.provider === "groq") {
        getE("groqKey").value = data.key;
        getE("groqModel").value = data.model;
        getE("groqAIPrompt").value = data.openai_prompt;
        getE("groqSpeechEndpoint").value = data.endpoint;
        getE("commandYes").checked = data.commands_enable
        getE("intentyes").checked = data.intentgraph
        getE("saveChatYes").checked = data.save_chat
      } else if (data.provider === "houndify") {
        getE("houndKey").value = data.key;
        getE("houndID").value = data.id;
      }
      checkKG();
    });
}

function setSTTLanguage() {
  const data = { language: getE("languageSelection").value };

  displayMessage("languageStatus", "Setting...");

  fetch("/api/set_stt_info", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("downloading")) {
        displayMessage("languageStatus", "Downloading model...");
        updateSTTLanguageDownload();
      } else {
        displayMessage("languageStatus", response);
        getE("languageSelectionDiv").style.display = response.includes("success") ? "block" : "none";
      }
    });
}

function updateSTTLanguageDownload() {

  const interval = setInterval(() => {
    fetch("/api/get_download_status")
      .then((response) => response.text())
      .then((response) => {
        displayMessage("languageStatus", response.includes("not downloading") ? "Initiating download..." : response)
        if (response.includes("success") || response.includes("error")) {
          displayMessage("languageStatus", response);
          getE("languageSelectionDiv").style.display = "block";
          clearInterval(interval);
        }
      });
  }, 500);
}

function sendRestart() {
  fetch("/api/reset")
    .then((response) => response.text())
    .then((response) => {
      displayMessage("restartStatus", response);
    });
}

function hideEditIntents() {
  getE("editIntentForm").style.display = "none";
  getE("editIntentStatus").innerHTML = "";
}

function showEditIntents() {
  getE("editIntentForm").style.display = "block";
}

function displayMessage(elementId, message) {
  const element = getE(elementId);
  element.innerHTML = "";
  const p = document.createElement("p");
  p.textContent = message;
  element.appendChild(p);
}

function displayError(elementId, message) {
  const element = getE(elementId);
  element.innerHTML = "";
  const error = document.createElement("p");
  error.innerHTML = message;
  element.appendChild(error);
}

function toggleSection(sectionToToggle, sectionToClose, foldableID) {
  const toggleSect = getE(sectionToToggle);
  const closeSect = getE(sectionToClose);

  if (toggleSect.style.display === "block") {
    closeSection(toggleSect, foldableID);
  } else {
    openSection(toggleSect, foldableID);
    closeSection(closeSect, foldableID);
  }
}

function openSection(sectionID) {
  sectionID.style.display = "block";
}

function closeSection(sectionID) {
  sectionID.style.display = "none";
}

function updateColor(id) {
  const l_id = id.replace("section", "icon");
  const elements = document.getElementsByName("icon");

  elements.forEach((element) => {
    element.classList.remove("selectedicon");
    element.classList.add("nowselectedicon");
  });

  const targetElement = document.getElementById(l_id);
  targetElement.classList.remove("notselectedicon");
  targetElement.classList.add("selectedicon");
}


function showLog() {
  toggleVisibility(["section-intents", "section-log", "section-botauth", "section-version", "section-uicustomizer"], "section-log", "icon-Logs");
  logDivArea = getE("botTranscriptedTextArea");
  getE("logscrollbottom").checked = true;
  logP = document.createElement("p");
  GetLog = true
  const interval = setInterval(() => {
    if (!GetLog) {
      clearInterval(interval);
      return;
    }
    const url = getE("logdebug").checked ? "/api/get_debug_logs" : "/api/get_logs";
    fetch(url)
      .then((response) => response.text())
      .then((logs) => {
        logDivArea.innerHTML = logs || "No logs yet, you must say a command to Vector. (this updates automatically)";
        if (getE("logscrollbottom").checked) {
          logDivArea.scrollTop = logDivArea.scrollHeight;
        }
      });
  }, 500);
}

function checkUpdate() {
  displayMessage("cVersion", "Checking for updates...");
  displayMessage("aUpdate", "");
  displayMessage("cCommit", "");
  fetch("/api/get_version_info")
    // type VersionInfo struct {
    // 	FromSource      bool   `json:"fromsource"`
    // 	InstalledVer    string `json:"installedversion"`
    // 	InstalledCommit string `json:"installedcommit"`
    // 	CurrentVer      string `json:"currentver"`
    // 	CurrentCommit   string `json:"currentcommit"`
    // 	UpdateAvailable bool   `json:"avail"`
    // }
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("error")) {
        // <p id="cVersion"></p>
        // <p style="display: none;" id="cCommit"></p>
        // <p id="aUpdate"></p>
        displayMessage(
          "cVersion",
          "There was an error: " + response
        );
        getE("updateGuideLink").style.display = "none";
      } else {
        const parsed = JSON.parse(response);
        if (parsed.fromsource) {
          if (!parsed.avail) {
            displayMessage("aUpdate", `You are on the latest version.`);
            getE("updateGuideLink").style.display = "none";
          } else {
            displayMessage("aUpdate", `A newer version of WirePod (commit: ${parsed.currentcommit}) is available! Use this guide to update WirePod: `);
            getE("updateGuideLink").style.display = "block";
          }
          displayMessage("cVersion", `Installed Commit: ${parsed.installedcommit}`);
        } else {
          displayMessage("cVersion", `Installed Version: ${parsed.installedversion}`);
          displayMessage("cCommit", `Based on wire-pod commit: ${parsed.installedcommit}`);
          getE("cCommit").style.display = "block";
          if (parsed.avail) {
            displayMessage("aUpdate", `A newer version of WirePod (${parsed.currentversion}) is available! Use this guide to update WirePod: `);
            getE("updateGuideLink").style.display = "block";
          } else {
            displayMessage("aUpdate", "You are on the latest version.");
            getE("updateGuideLink").style.display = "none";
          }
        }
      }
    });
}

function showSTT() {
  toggleVisibility(["section-weather", "section-restart", "section-kg", "section-language"], "section-language", "icon-STT");
  fetch("/api/get_stt_info")
    .then((response) => response.json())
    .then((parsed) => {
      // Set up the current STT service selection
      if (parsed.provider) {
        getE("sttServiceSelection").value = parsed.provider;
        updateSTTServiceOptions();
      }
      
      // Set up Whisper API settings if applicable
      if (parsed.provider === "whisper") {
        // Set provider (OpenAI, Groq, or custom)
        if (parsed.stt_provider) {
          // Determine if this is a custom provider
          if (parsed.stt_provider !== "openai" && parsed.stt_provider !== "groq") {
            getE("whisperProviderSelection").value = "custom";
          } else {
            getE("whisperProviderSelection").value = parsed.stt_provider;
          }
          updateWhisperProviderOptions();
        }
        
        // Set API key
        if (parsed.api_key) {
          getE("whisperApiKey").value = parsed.api_key;
        }
        
        // Set custom endpoint if using custom provider
        if (parsed.stt_provider === "custom" && parsed.endpoint) {
          getE("whisperEndpoint").value = parsed.endpoint;
        }
      }
      
      // Set up Whisper.cpp model if applicable
      if (parsed.provider === "whisper.cpp" && parsed.model) {
        getE("whisperCppModel").value = parsed.model;
      }
      
      // Set language
      if (parsed.language) {
        getE("languageSelection").value = parsed.language;
      }
    });
}

function updateSTTServiceOptions() {
  const service = getE("sttServiceSelection").value;
  
  // Hide all service-specific settings
  getE("whisperApiSettings").style.display = "none";
  getE("whisperCppSettings").style.display = "none";
  
  // Show relevant settings based on service
  if (service === "whisper") {
    getE("whisperApiSettings").style.display = "block";
    getE("languageSelectionDiv").style.display = "none"; // Hide language selection for Whisper API
    updateWhisperProviderOptions();
  } else if (service === "whisper.cpp") {
    getE("whisperCppSettings").style.display = "block";
    getE("languageSelectionDiv").style.display = "block"; // Show language selection
  } else {
    getE("languageSelectionDiv").style.display = "block"; // Show language selection
  }
}

function updateWhisperProviderOptions() {
  const provider = getE("whisperProviderSelection").value;
  
  // Hide all provider-specific elements first
  getE("whisperCustomEndpoint").style.display = "none";
  getE("openaiEndpointInfo").style.display = "none";
  getE("groqEndpointInfo").style.display = "none";
  getE("connectionStatus").innerHTML = ""; // Clear connection status
  
  // Show relevant elements based on provider
  if (provider === "openai") {
    getE("openaiEndpointInfo").style.display = "block";
  } else if (provider === "groq") {
    getE("groqEndpointInfo").style.display = "block";
  } else if (provider === "custom") {
    getE("whisperCustomEndpoint").style.display = "block";
  }
}

function saveSTTSettings() {
  const service = getE("sttServiceSelection").value;
  const language = getE("languageSelection").value;
  
  const data = {
    provider: service,
    language: language,
    stt_provider: "",
    api_key: "",
    endpoint: "",
    model: ""
  };
  
  // Add service-specific settings
  if (service === "whisper") {
    const whisperProvider = getE("whisperProviderSelection").value;
    data.stt_provider = whisperProvider;
    data.api_key = getE("whisperApiKey").value;
    
    // Set the endpoint based on provider
    if (whisperProvider === "openai") {
      data.endpoint = "https://api.openai.com/v1";
    } else if (whisperProvider === "groq") {
      data.endpoint = "https://api.groq.com/openai/v1";
    } else if (whisperProvider === "custom") {
      data.endpoint = getE("whisperEndpoint").value;
    }
  } else if (service === "whisper.cpp") {
    data.model = getE("whisperCppModel").value;
  }

  displayMessage("languageStatus", "Saving settings...");

  fetch("/api/set_stt_info", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("downloading")) {
        // Show progress bar for downloading
        getE("downloadProgress").style.display = "block";
        getE("languageStatus").innerHTML = "";
        startProgressMonitoring();
      } else {
        getE("downloadProgress").style.display = "none";
        displayMessage("languageStatus", response);
      }
    });
}

function startProgressMonitoring() {
  let progress = 0;
  const progressBar = getE("progressBar");
  const downloadStatus = getE("downloadStatus");
  
  // Simulate progress until we get real status
  const initialInterval = setInterval(() => {
    if (progress < 90) {
      progress += Math.random() * 5;
      progressBar.style.width = progress + "%";
    }
  }, 500);
  
  // Check actual download status
  const statusCheck = setInterval(() => {
    fetch("/api/get_download_status")
      .then(response => response.text())
      .then(status => {
        if (status.includes("success")) {
          clearInterval(initialInterval);
          clearInterval(statusCheck);
          progressBar.style.width = "100%";
          downloadStatus.textContent = "Download complete!";
          
          // Hide progress after 2 seconds and show success message
          setTimeout(() => {
            getE("downloadProgress").style.display = "none";
            displayMessage("languageStatus", "Model downloaded successfully. Settings saved.");
          }, 2000);
        } else if (status.includes("error")) {
          clearInterval(initialInterval);
          clearInterval(statusCheck);
          downloadStatus.textContent = "Error downloading model";
          progressBar.style.backgroundColor = "#f44336";
          
          // Show error message
          displayMessage("languageStatus", "Error: " + status);
        } else if (status.includes("not downloading")) {
          clearInterval(initialInterval);
          clearInterval(statusCheck);
          getE("downloadProgress").style.display = "none";
          displayMessage("languageStatus", "Settings saved successfully.");
        }
      });
  }, 1000);
}

function showVersion() {
  toggleVisibility(["section-log", "section-botauth", "section-intents", "section-version", "section-uicustomizer"], "section-version", "icon-Version");
  checkUpdate();
}

function showIntents() {
  toggleVisibility(["section-log", "section-botauth", "section-intents", "section-version", "section-uicustomizer"], "section-intents", "icon-Intents");
}

function showWeather() {
  toggleVisibility(["section-weather", "section-restart", "section-language", "section-kg"], "section-weather", "icon-Weather");
}

function showKG() {
  toggleVisibility(["section-weather", "section-restart", "section-language", "section-kg"], "section-kg", "icon-KG");
}

function toggleVisibility(sections, sectionToShow, iconId) {
  if (sectionToShow != "section-log") {
    GetLog = false;
  }
  sections.forEach((section) => {
    getE(section).style.display = "none";
  });
  getE(sectionToShow).style.display = "block";
  updateColor(iconId);
}

// Alias for backward compatibility
function showLanguage() {
  showSTT();
}

function testEndpointConnection() {
  const provider = getE("whisperProviderSelection").value;
  const apiKey = getE("whisperApiKey").value;
  let endpoint;
  
  // Get the appropriate endpoint based on provider
  if (provider === "openai") {
    endpoint = "https://api.openai.com/v1";
  } else if (provider === "groq") {
    endpoint = "https://api.groq.com/openai/v1";
  } else if (provider === "custom") {
    endpoint = getE("whisperEndpoint").value;
  }
  
  // Validate inputs
  if (!endpoint) {
    displayConnectionStatus("error", "Please enter a valid endpoint URL");
    return;
  }
  
  if (!apiKey) {
    displayConnectionStatus("error", "Please enter an API key");
    return;
  }
  
  // Show testing indicator
  displayConnectionStatus("testing", "Testing connection...");
  
  // Send request to test the endpoint
  fetch("/api/test_endpoint", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      endpoint: endpoint,
      api_key: apiKey
    }),
  })
    .then(response => response.json())
    .then(result => {
      if (result.status === "success") {
        displayConnectionStatus("success", "Connection successful!");
      } else if (result.status === "auth_error") {
        displayConnectionStatus("error", "Authentication failed: Invalid API key");
      } else {
        displayConnectionStatus("error", result.message);
      }
    })
    .catch(error => {
      displayConnectionStatus("error", "Request failed: " + error.message);
    });
}

function displayConnectionStatus(status, message) {
  const statusElem = getE("connectionStatus");
  statusElem.innerHTML = "";
  
  const statusIcon = document.createElement("span");
  statusIcon.style.marginRight = "5px";
  
  const statusText = document.createElement("span");
  statusText.textContent = message;
  
  if (status === "success") {
    statusIcon.innerHTML = "✅";
    statusText.style.color = "#4CAF50";
  } else if (status === "error") {
    statusIcon.innerHTML = "❌";
    statusText.style.color = "#f44336";
  } else if (status === "testing") {
    statusIcon.innerHTML = "🔄";
    statusText.style.color = "#2196F3";
  }
  
  statusElem.appendChild(statusIcon);
  statusElem.appendChild(statusText);
}