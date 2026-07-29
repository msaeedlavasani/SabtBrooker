import 'package:path/path.dart';
import 'package:sqflite/sqflite.dart';

class DatabaseHelper {
  static final DatabaseHelper _instance = DatabaseHelper._internal();
  static Database? _database;

  factory DatabaseHelper() => _instance;

  DatabaseHelper._internal();

  Future<Database> get database async {
    if (_database != null) return _database!;
    _database = await _initDatabase();
    return _database!;
  }

  Future<Database> _initDatabase() async {
    String path = join(await getDatabasesPath(), 'surveyor_db.db');
    return await openDatabase(
      path,
      version: 1,
      onCreate: (db, version) async {
        await db.execute('''
          CREATE TABLE surveys (
            id TEXT PRIMARY KEY,
            case_id TEXT,
            status TEXT,
            property_type TEXT,
            approx_area REAL,
            data_json TEXT,
            is_synced INTEGER DEFAULT 0,
            created_at TEXT
          )
        ''');
        await db.execute('''
          CREATE TABLE survey_photos (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            survey_id TEXT,
            file_path TEXT,
            side TEXT,
            latitude REAL,
            longitude REAL,
            created_at TEXT
          )
        ''');
      },
    );
  }

  Future<int> insertSurvey(Map<String, dynamic> row) async {
    Database db = await database;
    return await db.insert('surveys', row, conflictAlgorithm: ConflictAlgorithm.replace);
  }

  Future<int> insertPhoto(Map<String, dynamic> row) async {
    Database db = await database;
    return await db.insert('survey_photos', row);
  }

  Future<List<Map<String, dynamic>>> getUnsyncedSurveys() async {
    Database db = await database;
    return await db.query('surveys', where: 'is_synced = ?', whereArgs: [0]);
  }
}
