
export interface Field {
  key: string;
  label: string;
  type: 'text' | 'number' | 'date' | 'select' | 'checkbox' | 'textarea' | 'file' | 'repeat';
  options?: { v: string; l: string }[];
  required?: boolean;
  ref?: string;
  hint?: string;
  readonly?: boolean;
  visibleIf?: (data: any) => boolean;
  subFields?: Field[];
  addLabel?: string;
  validate?: (val: any, data: any) => string | null;
  chainNote?: string;
}

export interface Screen {
  id: string;
  title: string;
  note?: string;
  fields?: Field[];
  visibleIf?: (data: any) => boolean;
  kind?: 'otp' | 'result' | 'guidance';
  warn?: string;
  requireAck?: string;
  label?: string;
  prefix?: string;
  extra?: any;
}

const ROLE_OPTS = [
  { v: 'principal', l: 'اصیل' },
  { v: 'rep_natural', l: 'نمایندهٔ قانونی شخص حقیقی' },
  { v: 'rep_legal', l: 'نمایندهٔ قانونی شخص حقوقی' },
];

const SENA_OPTS = [
  { v: 'registered', l: 'ثبت‌نام کرده‌ام' },
  { v: 'not_registered', l: 'ثبت‌نام نکرده‌ام' },
];

const REP_METHOD_OPTS = [
  { v: 'document_upload', l: 'بارگذاری سند رسمی نمایندگی' },
  { v: 'legal_entity_db_inquiry', l: 'استعلام از پایگاه اشخاص حقوقی' },
];

function representativeFields(idPrefix: string): Field[] {
  return [
    {
      key: 'rep_verification_method', label: 'نحوهٔ احراز نمایندگی', type: 'select', options: REP_METHOD_OPTS,
      ref: 'ماده ۸', required: true,
      visibleIf: d => d.applicant_role === 'rep_legal'
    },
    {
      key: 'legal_person_national_id', label: 'شناسهٔ ملی شخص حقوقی', type: 'text',
      ref: 'ماده ۸', required: true,
      visibleIf: d => d.applicant_role === 'rep_legal'
    },
    {
      key: 'principal_national_code', label: 'کد ملی اصیل', type: 'text',
      ref: 'ماده ۱۰', required: true,
      visibleIf: d => d.applicant_role === 'rep_natural'
    },
    {
      key: 'principal_alive', label: 'اصیل در قید حیات است', type: 'checkbox',
      ref: 'ماده ۱۰', required: true,
      visibleIf: d => d.applicant_role === 'rep_natural'
    },
    {
      key: 'rep_doc_image_ref', label: 'تصویر سند رسمی نمایندگی', type: 'file',
      ref: 'ماده ۹/۱۰', required: true,
      visibleIf: d => (d.applicant_role === 'rep_legal' && d.rep_verification_method === 'document_upload') || d.applicant_role === 'rep_natural'
    },
    {
      key: 'rep_doc_date', label: 'تاریخ تنظیم سند نمایندگی', type: 'date',
      ref: 'ماده ۹/۱۰', required: true,
      visibleIf: d => (d.applicant_role === 'rep_legal' && d.rep_verification_method === 'document_upload') || d.applicant_role === 'rep_natural'
    },
    {
      key: 'rep_doc_official_id', label: 'شناسهٔ سند رسمی نمایندگی (اختیاری)', type: 'text',
      ref: 'ماده ۹/۱۰',
      visibleIf: d => (d.applicant_role === 'rep_legal' && d.rep_verification_method === 'document_upload') || d.applicant_role === 'rep_natural'
    },
    {
      key: 'rep_doc_verification_code', label: 'رمز تصدیق سند (اختیاری)', type: 'text',
      ref: 'ماده ۹/۱۰',
      visibleIf: d => (d.applicant_role === 'rep_legal' && d.rep_verification_method === 'document_upload') || d.applicant_role === 'rep_natural'
    },
  ];
}

export const MAP_SCREENS: Screen[] = [
  {
    id: 'identity', title: 'هویت و سمت متقاضی', note: 'ماده ۶',
    fields: [
      { key: 'applicant_national_code', label: 'کد ملی متقاضی', type: 'text', required: true },
      { key: 'applicant_phone', label: 'شمارهٔ تلفن همراه متقاضی', type: 'text', required: true, hint: 'باید در سامانهٔ شاهکار به‌نام متقاضی ثبت باشد' },
      { key: 'applicant_18_and_alive', label: 'متقاضی بالای ۱۸ سال و در قید حیات است', type: 'checkbox', required: true },
      { key: 'sena_status', label: 'وضعیت ثبت‌نام در سامانهٔ ثنا', type: 'select', options: SENA_OPTS, required: true },
      { key: 'applicant_role', label: 'سمت متقاضی', type: 'select', options: ROLE_OPTS, required: true },
      ...representativeFields('map')
    ]
  },
  {
    id: 'location_property', title: 'موقعیت مکانی و مشخصات ملک', note: 'ماده ۶',
    fields: [
      { key: 'province', label: 'استان', type: 'text', required: true },
      { key: 'county', label: 'شهرستان', type: 'text', required: true },
      { key: 'city_or_village', label: 'شهر/روستا', type: 'text', required: true },
      { key: 'geo_lat', label: 'عرض جغرافیایی ملک', type: 'number', required: true },
      { key: 'geo_lng', label: 'طول جغرافیایی ملک', type: 'number', required: true },
      { key: 'address', label: 'نشانی دقیق ملک', type: 'text', required: true },
      { key: 'postal_code', label: 'کدپستی (اختیاری)', type: 'text' },
      { key: 'property_type', label: 'نوع ملک', type: 'select', options: [{ v: 'land', l: 'زمین' }, { v: 'apartment', l: 'آپارتمان' }, { v: 'villa', l: 'ویلایی' }], required: true },
      { key: 'property_area_approx', label: 'مساحت تقریبی ملک (مترمربع)', type: 'number', required: true },
      { key: 'property_usage', label: 'کاربری ملک', type: 'select', options: [{ v: 'residential', l: 'مسکونی' }, { v: 'commercial', l: 'تجاری' }, { v: 'administrative', l: 'اداری' }, { v: 'agricultural', l: 'کشاورزی' }], required: true },
      { key: 'ownership_type', label: 'نوع مالکیت', type: 'select', options: [{ v: 'area', l: 'عرصه' }, { v: 'structure', l: 'اعیان' }, { v: 'area_and_structure', l: 'عرصه و اعیان' }], required: true },
      { key: 'has_structures', label: 'ملک دارای اعیانی است', type: 'checkbox' },
      { key: 'annex_count', label: 'تعداد منضمات (پارکینگ، انباری و...)', type: 'number' },
      { key: 'grant_third_party_access', label: 'درخواست اعطای دسترسی اشخاص دیگر به این نقشه را دارم', type: 'checkbox', ref: 'ماده ۶/۷' },
      {
        key: 'third_party_codes', label: 'فهرست کد ملی/شناسهٔ ملی اشخاص مجاز به دسترسی', type: 'repeat',
        ref: 'ماده ۷', visibleIf: d => d.grant_third_party_access, addLabel: 'افزودن شخص',
        subFields: [{ key: 'code', label: 'کد ملی یا شناسهٔ ملی', type: 'text' }]
      },
    ]
  },
  {
    id: 'expert_registry', title: 'تخصیص کارشناس امور ثبتی و حقوقی', note: 'ماده ۱۲',
    fields: [
      { key: 'registry_expert_national_code', label: 'کد ملی کارشناس امور ثبتی و حقوقی', type: 'text', ref: 'ماده مربوطه', required: true, hint: 'باید در فهرست کارشناسان مجاز موضوع مادهٔ ۴ دستورالعمل باشد' },
      { key: 'registry_expert_opinion', label: 'نظر کارشناس دربارهٔ احراز نمایندگی', type: 'select', options: [{ v: 'verified', l: 'نمایندگی احراز شد' }, { v: 'not_verified', l: 'نمایندگی احراز نشد' }], required: true },
      { key: 'registry_expert_notes', label: 'توضیحات کارشناس (اختیاری)', type: 'textarea' },
    ]
  },
  {
    id: 'surveyor_assignment', title: 'تخصیص کارشناس نقشه‌بردار', note: 'ماده ۱۳',
    fields: [{ key: 'surveyor_expert_national_code', label: 'کد ملی کارشناس نقشه‌بردار', type: 'text', required: true, hint: 'باید در فهرست کارشناسان مجاز موضوع مادهٔ ۵ دستورالعمل باشد و کد ملی‌اش در سامانهٔ مانا ثبت شده باشد' }]
  },
  { id: 'consent', kind: 'otp', title: 'رضایت با رمز یکبارمصرف' },
  {
    id: 'survey', title: 'نقشه‌برداری میدانی', note: 'ماده ۱۶/۱۷',
    fields: [
      { key: 'survey_visited_in_person', label: 'حضور شخصی کارشناس نقشه‌بردار در محل احراز شد', type: 'checkbox', required: true, ref: 'تبصرهٔ ماده ۱۶' },
      { key: 'photo_1', label: 'عکس ضلع اول ملک (Geo-tag)', type: 'file', required: true },
      { key: 'photo_2', label: 'عکس ضلع دوم ملک (Geo-tag)', type: 'file', required: true },
      { key: 'photo_3', label: 'عکس ضلع سوم ملک (Geo-tag)', type: 'file', required: true },
      { key: 'photo_4', label: 'عکس ضلع چهارم ملک (Geo-tag)', type: 'file', required: true },
      { key: 'survey_map_file_ref', label: 'فایل نقشهٔ ترسیم‌شده (AutoCAD/WebGIS)', type: 'file', required: true },
      { key: 'survey_descriptive_table_json', label: 'جدول اطلاعات توصیفی ملک (فرمت سامانهٔ مانا)', type: 'textarea', required: true },
      { key: 'applicant_approved_map', label: 'نقشه مورد تایید متقاضی قرار گرفت', type: 'checkbox', required: true, ref: 'ماده ۱۷' },
    ]
  },
  { id: 'final', kind: 'result', title: 'صدور کد رهگیری نقشه', label: 'کد رهگیری نقشهٔ ثبتی', prefix: 'MAP' },
];

export const CLAIM_SCREENS: Screen[] = [
    {
      id:'identity', title:'هویت و سمت متقاضی', note:'ماده ۶',
      fields:[
        {key:'applicant_national_code', label:'کد ملی متقاضی', type:'text', required:true},
        {key:'applicant_phone', label:'شمارهٔ تلفن همراه متقاضی', type:'text', required:true},
        {key:'applicant_18_and_alive', label:'متقاضی بالای ۱۸ سال و در قید حیات است', type:'checkbox', required:true},
        {key:'sena_status', label:'وضعیت ثبت‌نام در سامانهٔ ثنا', type:'select', options:SENA_OPTS, required:true},
        {key:'applicant_role', label:'سمت متقاضی', type:'select', options:ROLE_OPTS, required:true},
        ...representativeFields('claim')
      ]
    },
    {
      id:'map_ref', title:'اطلاع از تبصرهٔ ۵ و کد رهگیری نقشه', note:'ماده ۶',
      fields:[
        {key:'tabsare5_acknowledged', label:'از مفاد تبصرهٔ ۵ مادهٔ ۱۰ قانون (تبعات ادعای واهی) مطلع هستم', type:'checkbox', required:true},
        {key:'map_tracking_code', label:'کد رهگیری نقشهٔ ثبتی', type:'text', required:true, chainNote:'باید از سرویس «تهیه نقشه ثبتی» و دارای وضعیت صادرشده باشد'},
      ]
    },
    {
      id:'expert_registry', title:'تخصیص کارشناس امور ثبتی و حقوقی', note:'ماده ۱۱',
      fields:[
        {key:'registry_expert_national_code', label:'کد ملی کارشناس امور ثبتی و حقوقی', type:'text', ref:'ماده مربوطه', required:true, hint:'باید در فهرست کارشناسان مجاز موضوع مادهٔ ۴ دستورالعمل باشد'},
        {key:'registry_expert_opinion', label:'نظر کارشناس دربارهٔ احراز نمایندگی', type:'select', options:[{v:'verified',l:'نمایندگی احراز شد'},{v:'not_verified',l:'نمایندگی احراز نشد'}], required:true},
        {key:'registry_expert_notes', label:'توضیحات کارشناس (اختیاری)', type:'textarea'},
      ]
    },
    {id:'consent', kind:'otp', title:'رضایت و تایید هشدار قانونی', warn:'توجه: درج ادعای واهی مطابق تبصرهٔ ۵ مادهٔ ۱۰ قانون دارای تبعات قانونی است.', requireAck:'false_claim_warning_ack'},
    {
      id:'claim_details', title:'اطلاعات ملک و ادعا', note:'ماده ۱۴/۱۵/۱۶/۱۷',
      fields:[
        {key:'property_type', label:'نوع ملک', type:'select', options:[{v:'land',l:'زمین'},{v:'apartment',l:'آپارتمان'},{v:'villa',l:'ویلایی'}], required:true},
        {key:'property_usage', label:'کاربری ملک', type:'select', options:[{v:'residential',l:'مسکونی'},{v:'commercial',l:'تجاری'},{v:'administrative',l:'اداری'},{v:'agricultural',l:'کشاورزی'}], required:true},
        {key:'main_registry_plaque', label:'پلاک ثبتی اصلی (اختیاری)', type:'text'},
        {key:'sub_registry_plaque', label:'پلاک ثبتی فرعی (اختیاری)', type:'text'},
        {key:'claim_type', label:'نوع ادعا', type:'select', options:[
          {v:'ownership_of_object',l:'مالکیت عین'},{v:'right_of_easement',l:'حق ارتفاق'},
          {v:'right_of_usufruct',l:'حق انتفاع'},{v:'ownership_of_benefits',l:'مالکیت منافع'}], required:true},
        {key:'claimed_ownership_type', label:'نوع مالکیت مورد ادعا', type:'select', options:[{v:'area',l:'عرصه'},{v:'structure',l:'اعیان'},{v:'area_and_structure',l:'عرصه و اعیان'}], required:true},
        {key:'share_total', label:'سهم کل (مخرج کسر مالکیت)', type:'number', ref:'ماده ۱۷', required:true, visibleIf: d => d.claim_type === 'ownership_of_object'},
        {key:'share_partial', label:'سهم جزء (صورت کسر مالکیت)', type:'number', ref:'ماده ۱۷', required:true, visibleIf: d => d.claim_type === 'ownership_of_object',
          validate:(v,d)=> (d.claim_type==='ownership_of_object' && d.share_total && Number(v) > Number(d.share_total)) ? 'سهم جزء نباید بزرگ‌تر از سهم کل باشد.' : null },
        {key:'documents', label:'مستندات پشتیبان ادعا', type:'repeat', ref:'ماده ۱۶', required:true, minItems:1, addLabel:'افزودن مستند',
            subFields:[
                {key:'document_type', label:'نوع مستند', type:'select', options:[
                    {v:'sale_deed',l:'مبایعه‌نامه'},{v:'settlement_deed',l:'صلح‌نامه'},{v:'preliminary_contract',l:'قولنامه'},
                    {v:'partition_deed',l:'تقسیم‌نامه'},{v:'gift_deed',l:'هبه‌نامه'},{v:'court_ruling',l:'آرای محاکم'},
                    {v:'farmer_possession_record',l:'نسق زارعانه'},{v:'testimony_affidavit',l:'استشهادیه'},
                    {v:'inheritance_certificate',l:'گواهی انحصار وراثت'},{v:'other',l:'سایر'}], required:true},
                {key:'document_number', label:'شمارهٔ سند (اختیاری)', type:'text'},
                {key:'document_date', label:'تاریخ تنظیم سند (اختیاری)', type:'date'},
                {key:'document_image_ref', label:'تصویر سند', type:'file', required:true},
            ]},
        {key:'government_dues_applicable', label:'این ادعا شامل حقوق دولتی است', type:'checkbox'},
      ]
    },
    {id:'final', kind:'result', title:'صدور کد رهگیری ادعا', label:'کد رهگیری ادعا', prefix:'CLAIM', extra:{ govDuesGate:true } },
    {
      id:'guidance', title:'راهنمایی ثبتی', note:'ماده ۱۹', kind:'guidance',
      fields:[
        {key:'guidance_method', label:'روش ارائهٔ راهنمایی', type:'select', options:[{v:'expert',l:'کارشناس ثبتی و حقوقی'},{v:'ai_tool',l:'ابزار هوش مصنوعی'}], required:true},
      ]
    },
];

export const ACTION_SCREENS: Screen[] = [
    {
      id:'identity', title:'هویت و سمت متقاضی', note:'ماده ۶',
      fields:[
        {key:'applicant_national_code', label:'کد ملی متقاضی', type:'text', required:true},
        {key:'applicant_phone', label:'شمارهٔ تلفن همراه متقاضی', type:'text', required:true},
        {key:'applicant_18_and_alive', label:'متقاضی بالای ۱۸ سال و در قید حیات است', type:'checkbox', required:true},
        {key:'sena_status', label:'وضعیت ثبت‌نام در سامانهٔ ثنا', type:'select', options:SENA_OPTS, required:true},
        {key:'applicant_role', label:'سمت متقاضی', type:'select', options:ROLE_OPTS, required:true},
        ...representativeFields('action')
      ]
    },
    {
      id:'claim_ref', title:'ادعای مرجع و وضعیت فوت مدعی', note:'ماده ۶/۷',
      fields:[
        {key:'claim_tracking_code', label:'کد رهگیری درج ادعا', type:'text', required:true, chainNote:'نباید بیش از ۲ سال از تاریخ درج ادعا گذشته باشد؛ مگر با احتساب قاعدهٔ فوت مدعی'},
        {key:'claimant_deceased', label:'مدعی فوت شده است', type:'checkbox'},
        {key:'deceased_national_code', label:'کد ملی متوفی', type:'text', required:true, visibleIf: d => d.claimant_deceased, hint:'باید یکی از وراث طبق گواهی انحصار وراثت باشد'},
        {key:'death_date', label:'تاریخ فوت مدعی', type:'date', required:true, visibleIf: d => d.claimant_deceased, ref:'مهلت حداکثر ۵ ماه از این تاریخ (یا سقف ۲ سال، هرکدام دیرتر)'},
      ]
    },
    {
      id:'expert_registry', title:'تخصیص کارشناس امور ثبتی و حقوقی', note:'ماده ۱۲',
      fields:[
        {key:'registry_expert_national_code', label:'کد ملی کارشناس امور ثبتی و حقوقی', type:'text', ref:'ماده مربوطه', required:true, hint:'باید در فهرست کارشناسان مجاز موضوع مادهٔ ۴ دستورالعمل باشد'},
        {key:'registry_expert_opinion', label:'نظر کارشناس دربارهٔ احراز نمایندگی', type:'select', options:[{v:'verified',l:'نمایندگی احراز شد'},{v:'not_verified',l:'نمایندگی احراز نشد'}], required:true},
        {key:'registry_expert_notes', label:'توضیحات کارشناس (اختیاری)', type:'textarea'},
      ]
    },
    {id:'consent', kind:'otp', title:'رضایت و تایید نهایی'},
    {
      id:'certificate_details', title:'اطلاعات گواهی اقدام', note:'ماده ۱۵/۱۶',
      fields:[
        {key:'action_authority', label:'مرجع اقدام', type:'select', options:[
          {v:'registry_org',l:'سازمان ثبت'},{v:'judiciary',l:'قوهٔ قضاییه'},
          {v:'organizing_board',l:'هیئت سامان‌دهی'},{v:'determination_board',l:'هیئت تعیین تکلیف'}], required:true},
        {key:'action_type', label:'نوع اقدام', type:'select', options:[
          {v:'judicial_claim',l:'طرح دعوا در مراجع قضایی'},{v:'initial_registration',l:'ثبت اولیه'},
          {v:'organizing_board_request',l:'طرح تقاضا در هیئت سامان‌دهی'},{v:'determination_board_request',l:'طرح تقاضا در هیئت تعیین تکلیف'}], required:true},
        {key:'action_issue_date', label:'تاریخ صدور گواهی اقدام', type:'date', required:true},
        {key:'action_certificate_unique_id', label:'شناسهٔ یکتای گواهی اقدام (اختیاری)', type:'text'},
        {key:'action_certificate_image_ref', label:'تصویر گواهی اقدام', type:'file', required:true},
      ]
    },
    {id:'final', kind:'result', title:'صدور کد رهگیری نهایی', label:'کد رهگیری گواهی اقدام (سند نهایی زنجیره)', prefix:'ACT'},
];
