import 'dart:convert';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:geolocator/geolocator.dart';
import 'package:sabt_brooker_surveyor/models/case_model.dart';
import 'package:sabt_brooker_surveyor/utils/location_service.dart';
import 'package:sabt_brooker_surveyor/utils/database_helper.dart';

class SurveyDetailScreen extends StatefulWidget {
  final CaseModel caseData;
  const SurveyDetailScreen({super.key, required this.caseData});

  @override
  State<SurveyDetailScreen> createState() => _SurveyDetailScreenState();
}

class _SurveyDetailScreenState extends State<SurveyDetailScreen> {
  final LocationService _locationService = LocationService();
  final ImagePicker _picker = ImagePicker();
  final DatabaseHelper _dbHelper = DatabaseHelper();
  bool _saving = false;
  
  final Map<String, XFile?> _photos = {
    'North': null,
    'South': null,
    'East': null,
    'West': null,
  };

  final Map<String, Position?> _locations = {
    'North': null,
    'South': null,
    'East': null,
    'West': null,
  };

  Future<void> _capturePhoto(String side) async {
    try {
      Position position = await _locationService.getCurrentLocation();
      final XFile? photo = await _picker.pickImage(
        source: ImageSource.camera,
        imageQuality: 80,
      );

      if (photo != null) {
        setState(() {
          _photos[side] = photo;
          _locations[side] = position;
        });
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e')),
      );
    }
  }

  Future<void> _submitSurvey() async {
    setState(() => _saving = true);
    try {
      final surveyId = widget.caseData.id;
      
      await _dbHelper.insertSurvey({
        'id': surveyId,
        'case_id': widget.caseData.id,
        'status': 'captured',
        'created_at': DateTime.now().toIso8601String(),
        'data_json': jsonEncode({'city': widget.caseData.city}),
      });

      for (var entry in _photos.entries) {
        if (entry.value != null) {
          await _dbHelper.insertPhoto({
            'survey_id': surveyId,
            'file_path': entry.value!.path,
            'side': entry.key.toLowerCase(),
            'latitude': _locations[entry.key]!.latitude,
            'longitude': _locations[entry.key]!.longitude,
            'created_at': DateTime.now().toIso8601String(),
          });
        }
      }

      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('اطلاعات با موفقیت در دیتابیس محلی ذخیره شد (آفلاین)')),
      );
      Navigator.pop(context);
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('خطا در ذخیره‌سازی: $e')),
      );
    } finally {
      setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('بازدید: ${widget.caseData.city}'),
        backgroundColor: const Color(0xFF101B33),
        foregroundColor: Colors.white,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildInfoCard(),
            const SizedBox(height: 24),
            const Text('عکس‌های چهار طرف ملک (Geo-tagged)', 
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
            const SizedBox(height: 16),
            GridView.count(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              crossAxisCount: 2,
              mainAxisSpacing: 16,
              crossAxisSpacing: 16,
              children: [
                _buildPhotoSlot('North', 'ضلع شمال'),
                _buildPhotoSlot('South', 'ضلع جنوب'),
                _buildPhotoSlot('East', 'ضلع شرق'),
                _buildPhotoSlot('West', 'ضلع غرب'),
              ],
            ),
            const SizedBox(height: 32),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _photos.values.any((e) => e == null) || _saving ? null : _submitSurvey,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF101B33),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
                child: _saving 
                  ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
                  : const Text('ثبت و ذخیره در موبایل (Offline Save)', style: TextStyle(fontWeight: FontWeight.bold)),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey[100],
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey[300]!),
      ),
      child: Column(
        children: [
          _infoRow('شناسه پرونده', widget.caseData.id.substring(0, 8)),
          _infoRow('استان/شهر', '${widget.caseData.province}، ${widget.caseData.city}'),
          _infoRow('وضعیت', widget.caseData.status),
        ],
      ),
    );
  }

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.between,
        children: [
          Text(label, style: const TextStyle(color: Colors.grey, fontSize: 12)),
          Text(value, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
        ],
      ),
    );
  }

  Widget _buildPhotoSlot(String key, String label) {
    bool hasPhoto = _photos[key] != null;
    return GestureDetector(
      onTap: () => _capturePhoto(key),
      child: Container(
        decoration: BoxDecoration(
          color: hasPhoto ? Colors.green[50] : Colors.grey[50],
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: hasPhoto ? Colors.green : Colors.grey[300]!, width: 2),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            if (hasPhoto)
              Expanded(
                child: ClipRRect(
                  borderRadius: const BorderRadius.vertical(top: Radius.circular(14)),
                  child: Image.file(File(_photos[key]!.path), width: double.infinity, fit: CrossAxisAlignment.cover),
                ),
              )
            else
              const Icon(Icons.camera_alt_outlined, size: 32, color: Colors.grey),
            Padding(
              padding: const EdgeInsets.all(8.0),
              child: Column(
                children: [
                  Text(label, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                  if (hasPhoto)
                    Text('${_locations[key]!.latitude.toStringAsFixed(4)}, ${_locations[key]!.longitude.toStringAsFixed(4)}', 
                      style: const TextStyle(fontSize: 9, color: Colors.green, fontWeight: FontWeight.bold)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
