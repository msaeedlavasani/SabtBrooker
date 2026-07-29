import 'package:flutter/material.dart';
import 'package:sabt_brooker_surveyor/screens/home_screen.dart';

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
      home: const HomeScreen(),
    );
  }
}
