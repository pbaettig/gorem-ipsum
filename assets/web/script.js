// const healthToggleBtn = document.getElementById('health-toggle');
const healthCheckConfigForm = document.getElementById('health-config-form');
const healthCheckConfigFailEvery = document.getElementById(
  'health-config-failevery'
);
const healthCheckConfigFailRatio = document.getElementById(
  'health-config-failratio'
);
const healthCheckConfigFailSeq = document.getElementById(
  'health-config-failseq'
);
const healthCheckConfigFailEveryBtn = document.getElementById(
  'health-config-failevery-set'
);
const healthCheckConfigFailRatioBtn = document.getElementById(
  'health-config-failratio-set'
);
const healthCheckConfigFailSeqBtn = document.getElementById(
  'health-config-failseq-set'
);

const healthStatusDiv = document.querySelector('.health-status');

const httpGetBtn = document.getElementById('http-get');
const httpUrlTxt = document.getElementById('http-url');
const httpContentDiv = document.querySelector('.http-content');
const healthHistoryUl = document.getElementById('health-history-list');
const httpForm = document.getElementById('http-form');

var lastStatus = null;

function resetHealthcheckConfigForm() {
  fetch('/api/config/health?clear', {
    headers: new Headers({
      Authorization: 'Basic ' + btoa('username:password')
    })
  })
    .then(r => r.json())
    .then(d => {
      console.log(d);
      healthCheckConfigFailEvery.value = d['FailEvery'];
      healthCheckConfigFailRatio.value = d['FailRatio'];
      healthCheckConfigFailSeq.value = d['FailSeq'];
    });
}

function setHealthcheckConfig(key, value) {
  fetch(`/api/config/health?${key}=${value}`, {
    headers: new Headers({
      Authorization: 'Basic ' + btoa('username:password')
    })
  })
    .then(r => r.json())
    .then(d => {
      console.log(d);
      healthCheckConfigFailEvery.value = d['FailEvery'];
      healthCheckConfigFailRatio.value = d['FailRatio'];
      healthCheckConfigFailSeq.value = d['FailSeq'];
    });
}

function addHealthHistory(status) {
  let maxElements = 10;
  var li = document.createElement('li');
  if (status) {
    li.innerText = `✓ (${timestamp()})`;
    li.className = 'health-history-up';
  } else {
    li.innerText = `✗ (${timestamp()})`;
    li.className = 'health-history-down';
  }
  if (healthHistoryUl.children.length >= maxElements) {
    healthHistoryUl.removeChild(healthHistoryUl.children[maxElements - 1]);
  }

  healthHistoryUl.insertBefore(li, healthHistoryUl.children[0]);
}

function setHealthStatus(status) {
  let cl = healthStatusDiv.classList;
  let statusTxt = healthStatusDiv.children[0];

  if (status) {
    cl.remove('down');
    cl.add('up');
    statusTxt.innerText = `UP ✓ (last check: ${timestamp()})`;
  } else {
    cl.remove('up');
    cl.add('down');
    statusTxt.innerText = `DOWN ✗ (last check: ${timestamp()})`;
  }
  if (lastStatus != status) {
    addHealthHistory(status);
  }
}

function timestamp() {
  return new Date().toLocaleTimeString();
}

function updateHealthStatus() {
  let cl = healthStatusDiv.classList;
  let status = healthStatusDiv.children[0];
  fetch('/api/health').then(response => {
    setHealthStatus(response.ok);
    lastStatus = response.ok;
  });
}

function updateUI() {
  updateHealthStatus();
}

function toggleHealthStatus() {
  let cl = healthStatusDiv.classList;
  let p = healthStatusDiv.children[0];

  if (cl.contains('up')) {
    cl.remove('up');
    cl.add('down');
    p.innerText = 'DOWN ✗';
    return;
  }
  if (cl.contains('down')) {
    cl.remove('down');
    cl.add('up');
    p.innerText = 'UP ✓';
  }
}

// healthToggleBtn.addEventListener('click', function(e) {
//   toggleHealthStatus();
// });

console.log(httpForm);
httpForm.addEventListener('submit', function(e) {
  e.preventDefault();

  if (httpUrlTxt.value === '') {
    return;
  }
  httpContentDiv.innerHTML = '';
  console.log(httpUrlTxt.value);
  fetch(httpUrlTxt.value)
    .then(response => response.text())
    .then(r => {
      httpContentDiv.innerHTML = r;
    });
});

healthCheckConfigForm.addEventListener('submit', function(e) {
  console.log('submit');
  e.preventDefault();
});

healthCheckConfigForm.addEventListener('reset', function(e) {
  console.log('reset');
  e.preventDefault();
  resetHealthcheckConfigForm();
});

healthCheckConfigFailEveryBtn.addEventListener('click', function(e) {
  console.log(`FailEvery: ${healthCheckConfigFailEvery.value}`);
  setHealthcheckConfig('failevery', healthCheckConfigFailEvery.value);
});
healthCheckConfigFailRatioBtn.addEventListener('click', function(e) {
  console.log(`FailRatio: ${healthCheckConfigFailRatio.value}`);
  setHealthcheckConfig('failratio', healthCheckConfigFailRatio.value);
});
healthCheckConfigFailSeqBtn.addEventListener('click', function(e) {
  console.log(`FailSeq: ${healthCheckConfigFailSeq.value}`);
  setHealthcheckConfig('failseq', healthCheckConfigFailSeq.value);
});

updateUI();
window.setInterval(updateUI, 2000);
