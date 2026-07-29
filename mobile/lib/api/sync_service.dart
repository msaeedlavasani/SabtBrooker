import 'dart:convert';
import 'dart:io';
import 'package:dio/dio.dart';
import 'package:sabt_brooker_surveyor/api/api_client.dart';
import 'package:sabt_brooker_surveyor/utils/database_helper.dart';
import 'package:path/path.dart' as p;

class SyncService {
  final ApiClient _api = ApiClient();
  final DatabaseHelper _dbHelper = DatabaseHelper();
  final Dio _uploadDio = Dio();

  Future<void> syncData() async {
    final unsynced = await _dbHelper.getUnsyncedSurveys();
    
    for (var survey in unsynced) {
      try {
        final surveyId = survey['id'];
        
        // 1. Get photos for this survey
        final db = await _dbHelper.database;
        final photos = await db.query('survey_photos', where: 'survey_id = ?', whereArgs: [surveyId]);
        
        List<Map<String, dynamic>> uploadedPhotos = [];

        for (var photo in photos) {
          final filePath = photo['file_path'];
          final fileName = p.basename(filePath);
          
          // A. Get Presigned URL
          final presignedRes = await _api.get("/storage/presigned-url?name=$fileName");
          final uploadUrl = presignedRes.data['upload_url'];
          final fileId = presignedRes.data['file_id'];

          // B. Upload actual file to MinIO
          final file = File(filePath);
          await _uploadDio.put(
            uploadUrl,
            data: file.openRead(),
            options: Options(
              headers: {
                Headers.contentLengthHeader: await file.length(),
                'Content-Type': 'image/jpeg',
              },
            ),
          );

          uploadedPhotos.add({
            'file_path': fileId,
            'side': photo['side'],
            'latitude': photo['latitude'],
            'longitude': photo['longitude'],
          });
        }

        // 2. Submit Fieldwork to Backend
        await _api.post("/map-services/$surveyId/fieldwork/submit", {
          'property_type': 'land', // Mock or get from survey data
          'approx_area_sqm': 0,
          'map_file_path': '',
          'map_format': 'dwg',
          'descriptive_table': jsonDecode(survey['data_json']),
          'photos': uploadedPhotos,
          'grant_access_to_others': false,
        });

        // 3. Mark as synced in local DB
        await db.update('surveys', {'is_synced': 1}, where: 'id = ?', whereArgs: [surveyId]);
        
      } catch (e) {
        print("Failed to sync survey ${survey['id']}: $e");
        rethrow;
      }
    }
  }
}
