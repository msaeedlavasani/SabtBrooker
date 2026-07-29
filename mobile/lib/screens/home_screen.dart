import 'package:flutter/material.dart';
import 'package:sabt_brooker_surveyor/api/api_client.dart';
import 'package:sabt_brooker_surveyor/models/case_model.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final ApiClient _api = ApiClient();
  List<CaseModel> _cases = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _fetchCases();
  }

  Future<void> _fetchCases() async {
    try {
      final response = await _api.get("/cases");
      final List data = response.data;
      setState(() {
        _cases = data.map((e) => CaseModel.fromJson(e)).toList();
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('پرونده‌های نقشه‌برداری', style: TextStyle(fontWeight: FontWeight.bold)),
        backgroundColor: const Color(0xFF101B33),
        foregroundColor: Colors.white,
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: _cases.length,
              itemBuilder: (context, index) {
                final c = _cases[index];
                return Card(
                  elevation: 2,
                  margin: const EdgeInsets.bottom(12),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  child: ListTile(
                    contentPadding: const EdgeInsets.all(16),
                    title: Text('${c.province} - ${c.city}', style: const TextStyle(fontWeight: FontWeight.bold)),
                    subtitle: Text(c.addressDetail ?? 'بدون نشانی دقیق'),
                    trailing: const Icon(Icons.chevron_right, color: Color(0xFFAD8A3C)),
                    onTap: () {
                      // Navigate to Detail and Start Fieldwork
                    },
                  ),
                );
              },
            ),
    );
  }
}
