const healthToggleBtn = document.getElementById('health-toggle');
const healthStatusDiv = document.getElementById('healthstatus');

function toggleHealthStatus() {
  if (healthStatusDiv.className == 'status-up') {
    healthStatusDiv.className = 'status-down';
    healthStatusDiv.innerText = 'DOWN';

    return;
  }
  if (healthStatusDiv.className == 'status-down') {
    healthStatusDiv.className = 'status-up';
    healthStatusDiv.innerText = 'UP';
  }
}

healthToggleBtn.addEventListener('click', function(e) {
  toggleHealthStatus();
});
