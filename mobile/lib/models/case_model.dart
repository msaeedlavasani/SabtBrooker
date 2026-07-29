class CaseModel {
  final String id;
  final String province;
  final String city;
  final String? addressDetail;
  final String status;

  CaseModel({
    required this.id,
    required this.province,
    required this.city,
    this.addressDetail,
    required this.status,
  });

  factory CaseModel.fromJson(Map<String, dynamic> json) {
    return CaseModel(
      id: json['id'],
      province: json['province'],
      city: json['city'],
      addressDetail: json['address_detail'],
      status: json['status'],
    );
  }
}
