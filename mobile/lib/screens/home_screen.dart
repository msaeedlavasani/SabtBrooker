import 'package:flutter/material.dart';
import 'package:sabt_brooker_surveyor/api/api_client.dart';
import 'package:sabt_brooker_surveyor/api/sync_service.dart';
import 'package:sabt_brooker_surveyor/models/case_model.dart';
import 'package:sabt_brooker_surveyor/screens/survey_detail_screen.dart';
import 'package:sabt_brooker_surveyor/utils/database_helper.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final ApiClient _api = ApiClient();
  final SyncService _syncService = SyncService();
  final DatabaseHelper _dbHelper = DatabaseHelper();
  
  List<CaseModel> _cases = [];
  int _unsyncedCount = 0;
  bool _loading = true;
  bool _syncing = false;

  @override
  void initState() {
    super.initState();
    _refreshData();
  }

  Future<void> _refreshData() async {
    setState(() => _loading = true);
    try {
      // Fetch online cases
      final response = await _api.get("/cases");
      final List data = response.data;
      
      // Fetch offline unsynced count
      final unsynced = await _dbHelper.getUnsyncedSurveys();
      
      setState(() {
        _cases = data.map((e) => CaseModel.fromJson(e)).toList();
        _unsyncedCount = unsynced.length;
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  Future<void> _handleSync() async {
    setState(() => _syncing = true);
    try {
      await _syncService.syncData();
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('همگام‌سازی با موفقیت انجام شد')),
      );
      _refreshData();
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('خطا در همگام‌سازی: $e')),
      );
    } finally {
      setState(() => _syncing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('پنل نقشه‌بردار', style: TextStyle(fontWeight: FontWeight.bold)),
        backgroundColor: const Color(0xFF101B33),
        foregroundColor: Colors.white,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loading ? null : _refreshData,
          ),
        ],
      ),
      body: Column(
        children: [
          if (_unsyncedCount > 0) _buildSyncBanner(),
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _cases.length,
                    itemBuilder: (context, index) {
                      final c = _cases[index];
                      return _buildCaseCard(c);
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildSyncBanner() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      color: const Color(0xFFAD8A3C).withOpacity(0.1),
      child: Row(
        children: [
          const Icon(Icons.cloud_off, color: Color(0xFFAD8A3C)),
          const SizedBox(width: 12),
          Expanded(
            child: Text('$_unsyncedCount پرونده آماده همگام‌سازی است', 
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Color(0xFFAD8A3C))),
          ),
          ElevatedButton(
            onPressed: _syncing ? null : _handleSync,
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFFAD8A3C),
              foregroundColor: Colors.white,
              visualDensity: VisualDensity.compact,
            ),
            child: _syncing 
              ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
              : const Text('سینک اکنون'),
          ),
        ],
      ),
    );
  }

  Widget _buildCaseCard(CaseModel c) {
    return Card(
      elevation: 2,
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ListTile(
        contentPadding: const EdgeInsets.all(16),
        title: Text('${c.province} - ${c.city}', style: const TextStyle(fontWeight: FontWeight.bold)),
        subtitle: Text(c.addressDetail ?? 'بدون نشانی دقیق'),
        trailing: const Icon(Icons.chevron_right, color: Color(0xFFAD8A3C)),
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => SurveyDetailScreen(caseData: c),
            ),
          ).then((_) => _refreshData());
        },
      ),
    );
  }
}
