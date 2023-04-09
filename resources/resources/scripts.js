function uploadConsensusSet() {
  var formElement = document.getElementById("consensusSetFile");
  var xhr = new XMLHttpRequest();
  xhr.upload.onprogress = function(evt) {
    var percent = Math.round(evt.loaded / evt.total * 100);
    if (percent > 98) {
      percent = 99
    }
    document.getElementById("popup_content").innerHTML = "Uplading Consensus (" + percent + "%)"
  }
  xhr.onload = function() {
    document.getElementById("popup_content").innerHTML = "Uplading Consensus (100%)"
    console.log('Upload completed successfully.');
    document.getElementById("reload").submit()
  }
  xhr.open("POST", formElement.action);
  xhr.send(new FormData(formElement));
}
var txHistoryPageLines = []
function populateTxHistoryPage(json, sessionID) {
  var txHistoryPageElement = document.getElementById("tx_history_page")
  if (typeof(txHistoryPageElement) == 'undefined' || txHistoryPageElement == null) {
    return
  }
  if (json.length === 0) {
    return
  }
  if (txHistoryPageLines.length !== 0 
    && txHistoryPageLines[0].transaction_id === json[0].lines.transaction_id 
    && txHistoryPageLines[0].confirmed == json[0].lines.confirmed) {
    return
  }
  txHistoryPage = json.lines
  var cacheBuster = self.crypto.randomUUID()
  var txHistoryPageHtml = `
<ul class="row">
  <h3 class="col-6 center no-wrap">
    Transaction ID
  </h3>
  <li class="col-6 center no-wrap">
    Date
  </li>
  <li class="col-6 center no-wrap">
    Amount
  </li>
  <li class="col-6 center no-wrap">
    Fee
  </li>
  <li class="col-6 center no-wrap">
    Confirmed
  </li>
  <li class="col-6 center no-wrap">
    Notes
    <form class="inline-block" action="/gui/importExportNotesForm?&CACHE_BUSTER;" method="post">
      <input type="hidden" name="session_id" value="`+sessionID+`">
      <button class="small-button" title="Import/Export notes" type="submit">⇅</button>
    </form>
  </li>
</ul>
`
  for (const line of txHistoryPage) {
    const note = localStorage.getItem('note-'+line.short_transaction_id)
    if (line.confirmed == "No") {
      txHistoryPageHtml = txHistoryPageHtml + `<ul class="row-gray">`
    } else {
      txHistoryPageHtml = txHistoryPageHtml + `<ul class="row">`
    }
    txHistoryPageHtml = txHistoryPageHtml + `
  <li class="col-6 center no-wrap white-underline pad-col">
    <a href="https://scprime.info/?search=${line.transaction_id}" target="_blank" class="nolines">
      ${line.short_transaction_id}
    </a>
  </li>
  <li class="col-6 center no-wrap white-underline pad-col">
    ${line.time}
  </li>
  <li class="col-6 center no-wrap white-underline pad-col">
    ${line.amount}
  </li>
  <li class="col-6 center no-wrap white-underline pad-col">
    ${line.fee}
  </li>
  <li class="col-6 center no-wrap white-underline pad-col">
    ${line.confirmed}
  </li>
  <li class="col-6 center no-wrap white-underline pad-col">
    <input type="text" onchange="localStorage.setItem('note-${line.short_transaction_id}', this.value)" value="${note!==null?note:''}">
  </li>
</ul>
`
  }
  txHistoryPageElement.innerHTML = txHistoryPageHtml
  var txHistoryPageCountElement = document.getElementById("tx_history_page_count")
  if (typeof(txHistoryPageCountElement) != 'undefined' && txHistoryPageCountElement != null) {
    txHistoryPageCountElement.innerHTML = json.total
  }
  var txHistoryPagesElement = document.getElementById("tx_history_pages")
  if (typeof(txHistoryPagesElement) == 'undefined' || txHistoryPagesElement == null) {
    return
  }
  var txHistoryPagesHtml = ""
  for (var i = 0; i < json.total; i++) {
    var selected = ""
    if (i + 1 == json.current) {
      selected = "selected"
    }
    txHistoryPagesHtml = txHistoryPagesHtml + `<option ${selected} value='${i + 1}'>${i + 1}</option>`
  }
  txHistoryPagesElement.innerHTML = txHistoryPagesHtml
}
function refreshTxHistoryPage(sessionID) {
  var storedTxHistoryPageLinesKey = sessionID + "_storedTxHistoryPageLines"
  var storedTxHistoryPageLines = JSON.parse(localStorage.getItem(storedTxHistoryPageLinesKey) || "null");
  if (storedTxHistoryPageLines != null) {
    populateTxHistoryPage(storedTxHistoryPageLines, sessionID)
  } else {
    for (var i = 0; i < localStorage.length; i++){
      let key = localStorage.key(i)
      if (key.endsWith("_storedTxHistoryPageLines")) {
        localStorage.removeItem(key)
      }
    }
  }
  var txHistoryPageElement = document.getElementById("tx_history_page")
  if (typeof(txHistoryPageElement) != 'undefined' && txHistoryPageElement != null) {
    var data = new FormData();
    data.append("session_id", sessionID)
    fetch("/api/txHistoryPage", {method: "POST", body: data})
      .then(response => response.json())
      .then(result => {
        populateTxHistoryPage(result, sessionID)
        localStorage.setItem(storedTxHistoryPageLinesKey, JSON.stringify(result));
        setTimeout(() => {refreshTxHistoryPage(sessionID);}, 60000); // 1 minute in milliseconds
      })
      .catch(error => {
        console.error("Error:", error);
        setTimeout(() => {refreshTxHistoryPage(sessionID);}, 1000); // 1 second in milliseconds
      })
  } else {
    setTimeout(() => {refreshTxHistoryPage(sessionID);}, 50);
  }
}
function refreshBlockHeight(sessionID) {
  if (document.getElementsByClassName('block_height').length > 0) {
    var data = new FormData();
    data.append("session_id", sessionID)
    fetch("/gui/blockHeight", {method: "POST", body: data})
      .then(response => response.json())
      .then(result => {
        var blockHeight = result[0]
        var status = result[1]
        var color = result[2]
        // Automatically refresh form to make GUI smoother.
        if (status === "Synchronized") {
          var refreshForm = document.getElementById("refreshForm")
          if (typeof(refreshForm) != 'undefined' && refreshForm != null) {
            refreshForm.submit()
          }
        }
        for (const element of document.getElementsByClassName("block_height")){
          element.innerHTML=blockHeight;
        }
        for (const element of document.getElementsByClassName("status")){
          element.innerHTML=status;
        }
        for (const element of document.getElementsByClassName("status")){
          element.className="status " + color
        }
        if (status === "Synchronized") {
          setTimeout(() => {refreshBlockHeight(sessionID);}, 60000); // 1 minute in milliseconds
        } else {
          setTimeout(() => {refreshBlockHeight(sessionID);}, 1000); // 1 second in milliseconds
        }
      })
      .catch(error => {
        console.error("Error:", error);
        setTimeout(() => {refreshBlockHeight(sessionID);}, 60000); // 1 minute in milliseconds
      })
  }
}
function refreshBalance(sessionID) {
  var balance = document.getElementById("balance");
  if (typeof(balance) != 'undefined' && balance != null) {
    var data = new FormData();
    data.append("session_id", sessionID)
    fetch("/gui/balance", {method: "POST", body: data})
      .then(response => response.json())
      .then(result => {
        for (const element of document.getElementsByClassName("confirmed")){
          element.innerHTML = result[0];
        }
        for (const element of document.getElementsByClassName("unconfirmed")){
          element.innerHTML = result[1];
        }
        for (const element of document.getElementsByClassName("spfa_funds")){
          element.innerHTML = result[2];
        }
        for (const element of document.getElementsByClassName("spfb_funds")){
          element.innerHTML = result[3];
        }
        var whaleSize = document.getElementById("whale_size")
        if (typeof(whaleSize) != 'undefined' && whaleSize != null) {
          whaleSize.innerHTML = "Whale Size: " + result[5];
        }
        var whaleSizeButton = document.getElementById("whale_size_button")
        if (typeof(whaleSizeButton) != 'undefined' && whaleSizeButton != null) {
          whaleSizeButton.value = "Whale Size: " + result[5];
        }
        setTimeout(() => {refreshBalance(sessionID);}, 60000); // 1 minute in milliseconds
      })
      .catch(error => {
        console.error("Error:", error);
        setTimeout(() => {refreshBalance(sessionID);}, 60000); // 1 minute in milliseconds
      })
  } else {
    setTimeout(() => {refreshBalance(sessionID);}, 50);
  }
}
function refreshBootstrapperProgress() {
  if (document.getElementsByClassName('bootstrapper-progress').length > 0) {
    fetch("/gui/bootstrapperProgress")
      .then(response => response.json())
      .then(result => {
        var status = result[0]
        // Autorefresh wallet to make onboarding smoother.
        if (status === "100%") {
          var refreshBootstrapper = document.getElementById("refreshBootstrapper")
          if (typeof(refreshBootstrapper) != 'undefined' && refreshBootstrapper != null) {
            refreshBootstrapper.submit()
          }
        }
        for (const element of document.getElementsByClassName("bootstrapper-progress")){
          element.innerHTML = status;
        }
        setTimeout(() => {refreshBootstrapperProgress();}, 1000); // 1 second in milliseconds
      })
      .catch(error => {
        console.error("Error:", error);
        setTimeout(() => {refreshBootstrapperProgress();}, 1000); // 1 second in milliseconds
      })
  } else {
    setTimeout(() => {refreshBootstrapperProgress();}, 50);
  }
}
function refreshConsensusBuilderProgress() {
  if (document.getElementsByClassName('consensus-builder-progress').length > 0) {
    fetch("/gui/consensusBuilderProgress")
      .then(response => response.json())
      .then(result => {
        var status = result[0]
        // Autorefresh wallet to make onboarding smoother.
        if (status === "100%") {
          var refreshConsensusBuilder = document.getElementById("refreshConsensusBuilder")
          if (typeof(refreshConsensusBuilder) != 'undefined' && refreshConsensusBuilder != null) {
            refreshConsensusBuilder.submit()
          }
        }
        for (const element of document.getElementsByClassName("consensus-builder-progress")){
          element.innerHTML = status;
        }
        setTimeout(() => {refreshConsensusBuilderProgress();}, 1000); // 1 second in milliseconds
      })
      .catch(error => {
        console.error("Error:", error);
        setTimeout(() => {refreshConsensusBuilderProgress();}, 1000); // 1 second in milliseconds
      })
  } else {
    setTimeout(() => {refreshConsensusBuilderProgress();}, 50);
  }
}
function shutdownServer() {
  fetch("/shutdownServer", {method: "POST"})
    .then(response => response.json())
    .then(result => {
      if (result[0] === "true") {
        shutdownNotice()
      }
    })
    .catch(error => {
      console.error("Error:", error);
    })
}
function shutdownNotice() {
  var shutdownNoticeHtml = `
    <div class="col-5 left top no-wrap">
      <div>
        <img class="scprime-logo" alt="ScPrime Web Wallet" src="/gui/logo.png"/>
      </div>
    </div>
    <div id="popup" class="popup center">
      <h2 class="uppercase">Shutdown Notice</h2>
      <div class="middle pad blue-dashed" id="popup_content">Wallet was shutdown. You can now close your browser.</div>
    </div>
    <div id="fade" class="fade"></div>
  `
  var contentElement = document.getElementById("content")
  if (typeof(contentElement) != 'undefined' && contentElement != null) {
    contentElement.innerHTML = shutdownNoticeHtml
  } else {
    document.body.innerHTML = shutdownNoticeHtml
  }
}
function copyToClipboard(textToCopy) {
  var temp = document.createElement("input");
  temp.type = "text";
  temp.value = textToCopy;
  document.body.appendChild(temp);
  temp.select();
  document.execCommand("Copy");
  document.body.removeChild(temp);
}
// returns a new random wallet seed
function newWalletSeed() {
  return wasmNewWalletSeed()
}
// returns the zero address from the seed
function addressFromSeed(seed) {
  return wasmAddressFromSeed(seed)
}
refreshBootstrapperProgress()
refreshConsensusBuilderProgress()

