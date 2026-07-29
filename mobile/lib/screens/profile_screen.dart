import 'package:flutter/material.dart';

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('پروفایل کارشناس')),
      body: Column(
        children: [
          const SizedBox(height: 24),
          const CircleAvatar(
            radius: 50,
            backgroundColor: Colors.blueGrey,
            child: Icon(Icons.person, size: 50, color: Colors.white),
          ),
          const SizedBox(height: 16),
          const Text(
            'کارشناس: محمد سعیدی',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const Text('کد نظام مهندسی: ۱۲۳۴۵۶۷۸'),
          const SizedBox(height: 32),
          const Divider(),
          ListTile(
            leading: const Icon(Icons.settings),
            title: const Text('تنظیمات اتصال به سرور'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () {},
          ),
          ListTile(
            leading: const Icon(Icons.sync),
            title: const Text('وضعیت همگام‌سازی آفلاین'),
            subtitle: const Text('آخرین سینک: ۱۰ دقیقه پیش'),
            onTap: () {},
          ),
          const Spacer(),
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: ElevatedButton.icon(
              onPressed: () {},
              icon: const Icon(Icons.logout),
              label: const Text('خروج از حساب کاربری'),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.red.withOpacity(0.1),
                foregroundColor: Colors.red,
              ),
            ),
          ),
          const Text('نسخه ۱.۰.۰ (تیر ۱۴۰۵)', style: TextStyle(color: Colors.grey, fontSize: 12)),
          const SizedBox(height: 12),
        ],
      ),
    );
  }
}
