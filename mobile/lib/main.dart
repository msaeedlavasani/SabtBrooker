import 'package:flutter/material.dart';
import 'package:sabt_brooker_surveyor/screens/home_screen.dart';
import 'package:sabt_brooker_surveyor/screens/dashboard_screen.dart';
import 'package:sabt_brooker_surveyor/screens/profile_screen.dart';

void main() {
  runApp(const SurveyorApp());
}

class SurveyorApp extends StatelessWidget {
  const SurveyorApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'SabtBrooker Surveyor',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF101B33)),
        useMaterial3: true,
        fontFamily: 'Vazirmatn',
      ),
      home: const MainScaffold(),
      builder: (context, child) {
        return Directionality(
          textDirection: TextDirection.rtl,
          child: child!,
        );
      },
    );
  }
}

class MainScaffold extends StatefulWidget {
  const MainScaffold({super.key});

  @override
  State<MainScaffold> createState() => _MainScaffoldState();
}

class _MainScaffoldState extends State<MainScaffold> {
  int _selectedIndex = 1; // Default to Dashboard

  final List<Widget> _pages = [
    const HomeScreen(),
    const DashboardScreen(),
    const ProfileScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
        index: _selectedIndex,
        children: _pages,
      ),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _selectedIndex,
        selectedItemColor: const Color(0xFF101B33),
        onTap: (index) {
          setState(() {
            _selectedIndex = index;
          });
        },
        items: const [
          BottomNavigationBarItem(icon: Icon(Icons.list_alt), label: 'پرونده‌ها'),
          BottomNavigationBarItem(icon: Icon(Icons.dashboard), label: 'میز کار'),
          BottomNavigationBarItem(icon: Icon(Icons.person), label: 'پروفایل'),
        ],
      ),
    );
  }
}
