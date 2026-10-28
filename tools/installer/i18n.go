package main

// Тексты установщика: русский, қазақша, English, o'zbekcha, кыргызча.
// Язык — по языку Windows, можно переключить кнопкой на первом окне.

type texts struct {
	Name, Title, Checking, CheckingSub                     string
	Install, Update, Reinstall, Cancel, Close, Done, Retry string
	Lang                                                   string // подпись кнопки языка — на этом языке
	WelcomeMain, WelcomeText, Requires, Size, NoAdmin      string
	UpdateMain, UpdateText, SameMain, SameText             string
	Working, Downloading, Plugin, FFmpeg, Installing       string
	Of, MB                                                 string
	DoneMain, DoneText, DoneUpdates, OpenGuide             string
	ClosePremiere, ClosePremiereText                       string
	ErrMain, ErrNet, ErrSig, ErrHash, ErrInstall           string
	UnMain, UnText, UnDone, UnBtn                          string
	TestChannel                                            string
}

var langs = []string{"ru", "kk", "en", "uz", "ky"}

var T = map[string]texts{
	"ru": {
		Name: "kök Edit", Title: "Установка kök Edit", Checking: "Проверяю последнюю версию…", CheckingSub: "Это займёт пару секунд.",
		Install: "Установить", Update: "Обновить", Reinstall: "Переустановить", Cancel: "Отмена", Close: "Закрыть", Done: "Готово", Retry: "Повторить",
		Lang:        "Язык: Русский",
		WelcomeMain: "kök Edit %s для Premiere Pro", WelcomeText: "Синхронизация, AutoCam, паузы, звук, субтитры на казахском и русском, нарезки и проверка перед сдачей.",
		Requires: "Нужен Premiere Pro 2022–2025 на Windows", Size: "Скачается ≈ %s МБ.", NoAdmin: "Права администратора не нужны.",
		UpdateMain: "Обновить kök Edit до %s", UpdateText: "Сейчас стоит версия %s. Настройки и вход в аккаунт сохранятся.",
		SameMain: "У вас уже последняя версия %s", SameText: "Ничего делать не нужно. «Переустановить» — если плагин работает с ошибками.",
		Working: "Устанавливаю…", Downloading: "Скачиваю %s", Plugin: "плагин", FFmpeg: "ffmpeg (обработка звука и видео)", Installing: "Устанавливаю файлы…",
		Of: "из", MB: "МБ",
		DoneMain: "Готово!", DoneText: "Откройте Premiere Pro → меню «Окно» → «Расширения» → kök Edit.", DoneUpdates: "Дальше обновления приходят прямо в плагине.", OpenGuide: "Открыть короткую инструкцию",
		ClosePremiere: "Закройте Premiere Pro", ClosePremiereText: "Сохраните проекты, закройте Premiere Pro и нажмите «Повторить».",
		ErrMain: "Не получилось установить", ErrNet: "Нет связи с интернетом или сервер недоступен. Проверьте интернет и нажмите «Повторить».", ErrSig: "Подпись версии не сходится — установка остановлена. Скачайте установщик заново с kok.blue/edit.", ErrHash: "Файл скачался с ошибкой. Нажмите «Повторить».", ErrInstall: "Не удалось записать файлы плагина: %s\n\nЗакройте Premiere Pro и папку плагина в Проводнике и нажмите «Повторить».",
		UnMain: "Удалить kök Edit?", UnText: "Плагин пропадёт из Premiere Pro. Настройки, расшифровки и логи в папке «kokEdit» останутся.", UnDone: "kök Edit удалён.", UnBtn: "Удалить",
		TestChannel: "Тестовая версия",
	},
	"kk": {
		Name: "kök Edit", Title: "kök Edit орнату", Checking: "Соңғы нұсқаны тексеріп жатырмын…", CheckingSub: "Бірнеше секунд кетеді.",
		Install: "Орнату", Update: "Жаңарту", Reinstall: "Қайта орнату", Cancel: "Болдырмау", Close: "Жабу", Done: "Дайын", Retry: "Қайталау",
		Lang:        "Тіл: Қазақша",
		WelcomeMain: "Premiere Pro үшін kök Edit %s", WelcomeText: "Синхрондау, AutoCam, үзілістер, дыбыс, қазақша және орысша субтитрлер, қысқа видеолар және тапсыру алдындағы тексеру.",
		Requires: "Windows жүйесінде Premiere Pro 2022–2025 керек", Size: "≈ %s МБ жүктеледі.", NoAdmin: "Әкімші құқығы қажет емес.",
		UpdateMain: "kök Edit-ті %s нұсқасына жаңарту", UpdateText: "Қазір %s нұсқасы тұр. Баптаулар мен аккаунтқа кіру сақталады.",
		SameMain: "Сізде соңғы нұсқа тұр: %s", SameText: "Ештеңе істеудің қажеті жоқ. Плагин қатемен істесе — «Қайта орнату».",
		Working: "Орнатып жатырмын…", Downloading: "Жүктеп жатырмын: %s", Plugin: "плагин", FFmpeg: "ffmpeg (дыбыс пен видеоны өңдеу)", Installing: "Файлдарды орнатып жатырмын…",
		Of: "/", MB: "МБ",
		DoneMain: "Дайын!", DoneText: "Premiere Pro ашыңыз → «Окно / Window» мәзірі → «Расширения / Extensions» → kök Edit.", DoneUpdates: "Бұдан әрі жаңартулар плагиннің өзінде келеді.", OpenGuide: "Қысқа нұсқаулықты ашу",
		ClosePremiere: "Premiere Pro-ны жабыңыз", ClosePremiereText: "Жобаларды сақтап, Premiere Pro-ны жабыңыз да, «Қайталау» батырмасын басыңыз.",
		ErrMain: "Орнату сәтсіз болды", ErrNet: "Интернет жоқ немесе сервер қолжетімсіз. Интернетті тексеріп, «Қайталау» батырмасын басыңыз.", ErrSig: "Нұсқаның қолтаңбасы сәйкес келмейді — орнату тоқтатылды. Орнатқышты kok.blue/edit сайтынан қайта жүктеңіз.", ErrHash: "Файл қатемен жүктелді. «Қайталау» батырмасын басыңыз.", ErrInstall: "Плагин файлдарын жазу мүмкін болмады: %s\n\nPremiere Pro мен плагин папкасын жауып, «Қайталау» батырмасын басыңыз.",
		UnMain: "kök Edit-ті жою керек пе?", UnText: "Плагин Premiere Pro-дан жойылады. «kokEdit» папкасындағы баптаулар, транскрипциялар мен логтар қалады.", UnDone: "kök Edit жойылды.", UnBtn: "Жою",
		TestChannel: "Сынақ нұсқасы",
	},
	"en": {
		Name: "kök Edit", Title: "kök Edit Setup", Checking: "Checking for the latest version…", CheckingSub: "This takes a couple of seconds.",
		Install: "Install", Update: "Update", Reinstall: "Reinstall", Cancel: "Cancel", Close: "Close", Done: "Done", Retry: "Retry",
		Lang:        "Language: English",
		WelcomeMain: "kök Edit %s for Premiere Pro", WelcomeText: "Sync, AutoCam, pause removal, audio cleanup, Kazakh and Russian subtitles, short clips and pre-delivery checks.",
		Requires: "Requires Premiere Pro 2022–2025 on Windows", Size: "Download size ≈ %s MB.", NoAdmin: "No administrator rights needed.",
		UpdateMain: "Update kök Edit to %s", UpdateText: "Version %s is installed now. Your settings and sign-in will be kept.",
		SameMain: "You already have the latest version, %s", SameText: "Nothing to do. Use “Reinstall” if the plugin isn't working properly.",
		Working: "Installing…", Downloading: "Downloading %s", Plugin: "plugin", FFmpeg: "ffmpeg (audio and video processing)", Installing: "Installing files…",
		Of: "of", MB: "MB",
		DoneMain: "All set!", DoneText: "Open Premiere Pro → Window menu → Extensions → kök Edit.", DoneUpdates: "From now on, updates arrive right inside the plugin.", OpenGuide: "Open the quick guide",
		ClosePremiere: "Close Premiere Pro", ClosePremiereText: "Save your projects, close Premiere Pro and click “Retry”.",
		ErrMain: "Installation failed", ErrNet: "No internet connection or the server is unavailable. Check your connection and click “Retry”.", ErrSig: "The version signature doesn't match — installation stopped. Download the installer again from kok.blue/edit.", ErrHash: "A file was downloaded with errors. Click “Retry”.", ErrInstall: "Couldn't write the plugin files: %s\n\nClose Premiere Pro and the plugin folder in Explorer, then click “Retry”.",
		UnMain: "Uninstall kök Edit?", UnText: "The plugin will be removed from Premiere Pro. Settings, transcripts and logs in the “kokEdit” folder will stay.", UnDone: "kök Edit has been removed.", UnBtn: "Uninstall",
		TestChannel: "Test version",
	},
	"uz": {
		Name: "kök Edit", Title: "kök Edit o‘rnatish", Checking: "So‘nggi versiya tekshirilmoqda…", CheckingSub: "Bir necha soniya ketadi.",
		Install: "O‘rnatish", Update: "Yangilash", Reinstall: "Qayta o‘rnatish", Cancel: "Bekor qilish", Close: "Yopish", Done: "Tayyor", Retry: "Qayta urinish",
		Lang:        "Til: O‘zbekcha",
		WelcomeMain: "Premiere Pro uchun kök Edit %s", WelcomeText: "Sinxronlash, AutoCam, pauzalarni olib tashlash, ovoz, qozoq va rus tilidagi subtitrlar, qisqa videolar va topshirishdan oldingi tekshiruv.",
		Requires: "Windows’da Premiere Pro 2022–2025 kerak", Size: "≈ %s MB yuklab olinadi.", NoAdmin: "Administrator huquqlari kerak emas.",
		UpdateMain: "kök Edit’ni %s versiyasiga yangilash", UpdateText: "Hozir %s versiyasi o‘rnatilgan. Sozlamalar va hisobga kirish saqlanadi.",
		SameMain: "Sizda eng so‘nggi versiya o‘rnatilgan: %s", SameText: "Hech narsa qilish shart emas. Plagin xato bilan ishlasa — «Qayta o‘rnatish».",
		Working: "O‘rnatilmoqda…", Downloading: "Yuklab olinmoqda: %s", Plugin: "plagin", FFmpeg: "ffmpeg (ovoz va videoni qayta ishlash)", Installing: "Fayllar o‘rnatilmoqda…",
		Of: "/", MB: "MB",
		DoneMain: "Tayyor!", DoneText: "Premiere Pro’ni oching → «Window» menyusi → «Extensions» → kök Edit.", DoneUpdates: "Endi yangilanishlar to‘g‘ridan-to‘g‘ri plaginda keladi.", OpenGuide: "Qisqa qo‘llanmani ochish",
		ClosePremiere: "Premiere Pro’ni yoping", ClosePremiereText: "Loyihalarni saqlang, Premiere Pro’ni yoping va «Qayta urinish» tugmasini bosing.",
		ErrMain: "O‘rnatib bo‘lmadi", ErrNet: "Internet yo‘q yoki server ishlamayapti. Internetni tekshirib, «Qayta urinish» tugmasini bosing.", ErrSig: "Versiya imzosi mos kelmadi — o‘rnatish to‘xtatildi. O‘rnatuvchini kok.blue/edit saytidan qayta yuklab oling.", ErrHash: "Fayl xato bilan yuklandi. «Qayta urinish» tugmasini bosing.", ErrInstall: "Plagin fayllarini yozib bo‘lmadi: %s\n\nPremiere Pro va plagin papkasini yopib, «Qayta urinish» tugmasini bosing.",
		UnMain: "kök Edit o‘chirilsinmi?", UnText: "Plagin Premiere Pro’dan olib tashlanadi. «kokEdit» papkasidagi sozlamalar, transkripsiyalar va loglar qoladi.", UnDone: "kök Edit o‘chirildi.", UnBtn: "O‘chirish",
		TestChannel: "Sinov versiyasi",
	},
	"ky": {
		Name: "kök Edit", Title: "kök Edit орнотуу", Checking: "Акыркы версия текшерилүүдө…", CheckingSub: "Бир нече секунд кетет.",
		Install: "Орнотуу", Update: "Жаңыртуу", Reinstall: "Кайра орнотуу", Cancel: "Жокко чыгаруу", Close: "Жабуу", Done: "Даяр", Retry: "Кайталоо",
		Lang:        "Тил: Кыргызча",
		WelcomeMain: "Premiere Pro үчүн kök Edit %s", WelcomeText: "Синхрондоо, AutoCam, тыныгууларды кесүү, үн, казакча жана орусча субтитрлер, кыска видеолор жана тапшыруу алдындагы текшерүү.",
		Requires: "Windows’то Premiere Pro 2022–2025 керек", Size: "≈ %s МБ жүктөлөт.", NoAdmin: "Администратордун укугу керек эмес.",
		UpdateMain: "kök Edit’ти %s версиясына жаңыртуу", UpdateText: "Азыр %s версиясы орнотулган. Жөндөөлөр жана аккаунтка кирүү сакталат.",
		SameMain: "Сизде акыркы версия орнотулган: %s", SameText: "Эч нерсе кылуунун кереги жок. Плагин ката менен иштесе — «Кайра орнотуу».",
		Working: "Орнотулууда…", Downloading: "Жүктөлүүдө: %s", Plugin: "плагин", FFmpeg: "ffmpeg (үн жана видеону иштетүү)", Installing: "Файлдар орнотулууда…",
		Of: "/", MB: "МБ",
		DoneMain: "Даяр!", DoneText: "Premiere Pro’ну ачыңыз → «Window / Окно» менюсу → «Extensions / Расширения» → kök Edit.", DoneUpdates: "Мындан ары жаңыртуулар түз эле плагинде келет.", OpenGuide: "Кыска колдонмону ачуу",
		ClosePremiere: "Premiere Pro’ну жабыңыз", ClosePremiereText: "Долбоорлорду сактап, Premiere Pro’ну жабыңыз жана «Кайталоо» баскычын басыңыз.",
		ErrMain: "Орнотуу ишке ашкан жок", ErrNet: "Интернет жок же сервер жеткиликсиз. Интернетти текшерип, «Кайталоо» баскычын басыңыз.", ErrSig: "Версиянын кол тамгасы дал келген жок — орнотуу токтотулду. Орноткучту kok.blue/edit сайтынан кайра жүктөп алыңыз.", ErrHash: "Файл ката менен жүктөлдү. «Кайталоо» баскычын басыңыз.", ErrInstall: "Плагиндин файлдарын жазуу мүмкүн болгон жок: %s\n\nPremiere Pro’ну жана плагиндин папкасын жаап, «Кайталоо» баскычын басыңыз.",
		UnMain: "kök Edit өчүрүлсүнбү?", UnText: "Плагин Premiere Pro’дон алынып салынат. «kokEdit» папкасындагы жөндөөлөр, транскрипциялар жана логдор калат.", UnDone: "kök Edit өчүрүлдү.", UnBtn: "Өчүрүү",
		TestChannel: "Сыноо версиясы",
	},
}

// язык Windows → наш язык
func detectLang(langID uint16) string {
	switch langID & 0x3ff {
	case 0x19:
		return "ru"
	case 0x3f:
		return "kk"
	case 0x43:
		return "uz"
	case 0x40:
		return "ky"
	case 0x09:
		return "en"
	case 0x22, 0x23, 0x2b, 0x2c, 0x28, 0x42, 0x37: // украинский, белорусский, армянский, азербайджанский, таджикский, туркменский, грузинский — чаще всего читают по-русски
		return "ru"
	}
	return "en"
}
