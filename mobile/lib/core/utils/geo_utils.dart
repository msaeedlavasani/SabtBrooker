import 'dart:io';
import 'package:geolocator/geolocator.dart';
import 'package:exif/exif.dart';

class GeoUtils {
  /// Check if GPS is enabled and permissions are granted
  static Future<bool> checkPermission() async {
    bool serviceEnabled;
    LocationPermission permission;

    serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) return false;

    permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) return false;
    }
    
    if (permission == LocationPermission.deniedForever) return false;

    return true;
  }

  /// Get current position
  static Future<Position> getCurrentPosition() async {
    return await Geolocator.getCurrentPosition(
      desiredAccuracy: LocationAccuracy.high,
    );
  }

  /// Verify if a file has valid EXIF GPS data
  static Future<Map<String, dynamic>?> getExifData(String filePath) async {
    final fileBytes = File(filePath).readAsBytesSync();
    final data = await decodeExifFromBytes(fileBytes);
    
    if (data.isEmpty) return null;
    
    return data;
  }
}
