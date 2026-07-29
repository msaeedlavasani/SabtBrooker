import 'package:flutter/material.dart';
import 'package:sabt_brooker_surveyor/api/api_client.dart';
import 'package:sabt_brooker_surveyor/models/case_model.dart';
import 'package:sabt_brooker_surveyor/screens/survey_detail_screen.dart';

class HomeScreen extends StatefulWidget {
...
                    trailing: const Icon(Icons.chevron_right, color: Color(0xFFAD8A3C)),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => SurveyDetailScreen(caseData: c),
                        ),
                      );
                    },
                  ),
                );
              },
            ),
    );
  }
}
