/**
 * Main Application JavaScript
 * Gym Equipment Booking System (Go Backend)
 */

document.addEventListener('DOMContentLoaded', () => {
  // Auto-dismiss alerts after 5 seconds
  const alerts = document.querySelectorAll('.alert-dismissible');
  alerts.forEach(alert => {
    setTimeout(() => {
      const bsAlert = bootstrap.Alert.getOrCreateInstance(alert);
      if (bsAlert) bsAlert.close();
    }, 5000);
  });

  // Confirm delete or cancel action dialogs
  document.querySelectorAll('[data-confirm]').forEach(el => {
    el.addEventListener('click', (e) => {
      const message = el.getAttribute('data-confirm') || 'คุณแน่ใจหรือไม่ว่าต้องการดำเนินการนี้?';
      if (!confirm(message)) {
        e.preventDefault();
      }
    });
  });
});
