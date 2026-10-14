# Implementation Plan: ระบบจองเข้าใช้อุปกรณ์ในยิม

## 1. Project Overview

สร้างเว็บ **ระบบจองเข้าใช้อุปกรณ์ในยิม (Gym Equipment Booking System)** สำหรับผู้ดูแลระบบ (Admin) และสมาชิกยิม (Member) เพื่อลดปัญหาการแย่งเครื่องเล่นและช่วยจัดสรรเวลาการใช้งาน

## 2. Technology Stack

* Frontend: HTML + CSS + Bootstrap + JavaScript
* Backend: Go
* Database: MySQL / MariaDB
* Security: Owasp Top 10

---

## 3. User Roles

### Admin (ผู้ดูแลระบบ)

* จัดการหมวดหมู่อุปกรณ์ (เช่น โซนคาร์ดิโอ, โซนเวทเทรนนิ่ง)
* เพิ่ม/แก้ไข/ลบ ข้อมูลอุปกรณ์และเครื่องเล่น
* ปรับสถานะอุปกรณ์ (เช่น พร้อมใช้งาน, ส่งซ่อม)
* ดูตารางการจองอุปกรณ์ทั้งหมด
* ยกเลิกการจอง (กรณีฉุกเฉินหรือเครื่องเสีย)
* ดูรายชื่อและประวัติการใช้งานของสมาชิก

### Member (สมาชิกยิม)

* ดูรายการอุปกรณ์และสถานะว่าง/ไม่ว่าง แบบ Real-time
* จองเวลาเข้าใช้อุปกรณ์ล่วงหน้า
* ยกเลิกการจองของตนเอง
* ดูประวัติการจองและสถิติการใช้งานของตนเอง

---

## 4. Core Features

### Authentication

* Register (สมัครสมาชิก)
* Login (เข้าสู่ระบบ)
* Logout (ออกจากระบบ)
* Password Hashing (เข้ารหัสผ่านเพื่อความปลอดภัย)
* Session Management (จัดการเซสชันการล็อกอิน)

### Dashboard

* **Admin Dashboard:** แสดงจำนวนสมาชิกทั้งหมด, สถิติเครื่องเล่นที่ถูกจองมากที่สุด, รายการแจ้งซ่อม, และคิวการจองในวันนี้
* **Member Dashboard:** แสดงรายการจองที่กำลังจะมาถึง (Upcoming Bookings), อุปกรณ์ที่จองบ่อย, และปุ่มทางลัดสำหรับจองอุปกรณ์

### Booking Management

* ปฏิทิน/ตารางเวลา สำหรับเลือกช่วงเวลาจอง
* ระบบตรวจสอบเวลาซ้ำซ้อน (Prevent Double Booking)

---

## 5. Suggested Database

users
- id
- name
- email
- password_hash
- role (admin, member)
- phone

categories
- id
- name (เช่น คาร์ดิโอ, ฟรีเวท, แมชชีน)

equipments
- id
- category_id
- name
- description
- image_path
- status (available, maintenance)

bookings
- id
- user_id
- equipment_id
- start_time (วันและเวลาที่เริ่ม)
- end_time (วันและเวลาที่สิ้นสุด)
- status (pending, confirmed, cancelled)
- created_at

---

## Relationship หลัก:

Gym System
  |
  |-- Categories (หมวดหมู่เครื่องเล่น)
  |     |
  |     |-- Equipments (อุปกรณ์/เครื่องเล่น)
  |           |
  |           |-- Bookings (ประวัติการจองของอุปกรณ์นี้)
  |
  |-- Users (สมาชิกยิม / Admin)
        |
        |-- Bookings (รายการจองของสมาชิกคนนี้)