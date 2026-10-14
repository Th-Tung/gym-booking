/**
 * Booking Slot Picker & Live Availability Checker
 * Gym Equipment Booking System (Go Backend)
 */

document.addEventListener('DOMContentLoaded', () => {
  const dateInput = document.getElementById('booking_date');
  const equipmentSelect = document.getElementById('equipment_id');
  const slotsContainer = document.getElementById('slots_container');
  const startTimeInput = document.getElementById('start_time');
  const endTimeInput = document.getElementById('end_time');
  const selectedSlotDisplay = document.getElementById('selected_slot_display');
  const submitBtn = document.getElementById('btn_submit_booking');

  if (!slotsContainer || !dateInput) return;

  // Operating Hours: 06:00 to 22:00 in 2-hour increments
  const timeSlots = [
    { start: '06:00', end: '08:00' },
    { start: '08:00', end: '10:00' },
    { start: '10:00', end: '12:00' },
    { start: '12:00', end: '14:00' },
    { start: '14:00', end: '16:00' },
    { start: '16:00', end: '18:00' },
    { start: '18:00', end: '20:00' },
    { start: '20:00', end: '22:00' }
  ];

  async function loadSlots() {
    const eqId = equipmentSelect ? equipmentSelect.value : (dateInput.dataset.equipmentId || '');
    const date = dateInput.value;

    if (!eqId || !date) {
      slotsContainer.innerHTML = '<div class="col-12 text-muted text-center py-4">กรุณาเลือกอุปกรณ์และวันที่</div>';
      return;
    }

    slotsContainer.innerHTML = '<div class="col-12 text-center py-4 text-muted"><div class="spinner-border spinner-border-sm text-success me-2"></div> กำลังตรวจสอบช่วงเวลาว่าง...</div>';

    try {
      const res = await fetch(`/api/availability?equipment_id=${encodeURIComponent(eqId)}&date=${encodeURIComponent(date)}`);
      const data = await res.json();

      if (!data.success) {
        slotsContainer.innerHTML = `<div class="col-12 text-danger text-center py-3">${data.message || 'เกิดข้อผิดพลาดในการโหลดข้อมูล'}</div>`;
        return;
      }

      const bookedRanges = data.booked_slots || [];
      const isEquipmentMaintenance = data.equipment && data.equipment.status === 'maintenance';

      if (isEquipmentMaintenance) {
        slotsContainer.innerHTML = `<div class="col-12 text-danger text-center py-4"><i class="bi bi-exclamation-triangle-fill fs-3 d-block mb-2"></i> อุปกรณ์นี้กำลังอยู่ระหว่างส่งซ่อมบำรุง (Maintenance) ไม่สามารถจองได้</div>`;
        if (submitBtn) submitBtn.disabled = true;
        return;
      }

      renderSlots(date, bookedRanges);
    } catch (err) {
      console.error(err);
      slotsContainer.innerHTML = '<div class="col-12 text-danger text-center py-3">เกิดข้อผิดพลาดในการเชื่อมต่อเซิร์ฟเวอร์</div>';
    }
  }

  function renderSlots(dateStr, bookedRanges) {
    slotsContainer.innerHTML = '';
    const now = new Date();
    const isToday = (dateStr === now.toISOString().split('T')[0]);

    // Reset current selection
    if (startTimeInput) startTimeInput.value = '';
    if (endTimeInput) endTimeInput.value = '';
    if (selectedSlotDisplay) selectedSlotDisplay.textContent = 'ยังไม่ได้เลือกช่วงเวลา';
    if (submitBtn) submitBtn.disabled = true;

    timeSlots.forEach(slot => {
      const slotStart = new Date(`${dateStr}T${slot.start}:00`);
      const slotEnd = new Date(`${dateStr}T${slot.end}:00`);

      // Check if slot is already in the past
      const isPast = isToday && (slotStart < now);

      // Check collision with any existing booked range (Strict Local Time)
      const isBooked = bookedRanges.some(b => {
        const cleanStart = String(b.start_time).replace(' ', 'T').replace('Z', '');
        const cleanEnd = String(b.end_time).replace(' ', 'T').replace('Z', '');
        const bStart = new Date(cleanStart);
        const bEnd = new Date(cleanEnd);
        return (slotStart < bEnd && slotEnd > bStart);
      });

      const col = document.createElement('div');
      col.className = 'col-6 col-md-4 col-lg-3';

      const btn = document.createElement('div');
      btn.className = 'slot-btn';

      if (isBooked) {
        btn.classList.add('booked');
        btn.innerHTML = `<i class="bi bi-x-circle me-1"></i> ${slot.start} - ${slot.end}<br><small class="text-danger">ถูกใช้งาน</small>`;
      } else if (isPast) {
        btn.classList.add('booked');
        btn.style.opacity = '0.4';
        btn.innerHTML = `<i class="bi bi-clock-history me-1"></i> ${slot.start} - ${slot.end}<br><small class="text-secondary">เวลาผ่านไปแล้ว</small>`;
      } else {
        btn.innerHTML = `<i class="bi bi-check2-circle text-success me-1"></i> ${slot.start} - ${slot.end}<br><small class="text-success">ว่าง</small>`;
        btn.addEventListener('click', () => {
          document.querySelectorAll('.slot-btn').forEach(b => b.classList.remove('selected'));
          btn.classList.add('selected');

          const formattedStart = `${dateStr} ${slot.start}:00`;
          const formattedEnd = `${dateStr} ${slot.end}:00`;

          if (startTimeInput) startTimeInput.value = formattedStart;
          if (endTimeInput) endTimeInput.value = formattedEnd;
          if (selectedSlotDisplay) {
            selectedSlotDisplay.innerHTML = `<strong class="text-success">${slot.start} - ${slot.end} น.</strong> <span class="badge bg-success bg-opacity-25 text-success ms-2">รอบ 2 ชั่วโมง</span> (วันที่ ${dateStr})`;
          }
          if (submitBtn) submitBtn.disabled = false;
        });
      }

      col.appendChild(btn);
      slotsContainer.appendChild(col);
    });
  }

  // Event Listeners
  if (dateInput) {
    dateInput.addEventListener('change', loadSlots);
  }
  if (equipmentSelect) {
    equipmentSelect.addEventListener('change', loadSlots);
  }

  // Initial load
  loadSlots();
});
