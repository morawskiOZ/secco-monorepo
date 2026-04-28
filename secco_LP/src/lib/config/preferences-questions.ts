import type { FormConfig } from './questions';

export const preferencesConfig: FormConfig = {
	welcome: {
		title: 'Ankieta preferencji',
		text: 'Wypełnij ankietę, abyśmy mogli lepiej poznać Twoje potrzeby i preferencje przed projektem.'
	},

	questions: [
		{
			id: 'p1',
			number: 1,
			text: 'Kto wypełnia ankietę?',
			type: 'contact_form',
			required: true,
			fields: [
				{ id: 'first_name', label: 'Imię', type: 'text', required: true, placeholder: 'Jan' },
				{ id: 'last_name', label: 'Nazwisko', type: 'text', placeholder: 'Kowalski' },
				{ id: 'phone', label: 'Numer telefonu', type: 'tel', placeholder: '+48 512 345 678' },
				{ id: 'email', label: 'Email', type: 'email', required: true, placeholder: 'name@example.com' }
			]
		},
		{
			id: 'p2',
			number: 2,
			text: 'Ile osób na co dzień będzie użytkować projektowane mieszkanie?',
			type: 'text'
		},
		{
			id: 'p3',
			number: 3,
			text: 'Czy domownicy są alergikami? Jeśli tak, to na co?',
			type: 'long_text'
		},
		{
			id: 'p4',
			number: 4,
			text: 'Czy któryś z domowników pracuje w domu? Jeśli tak, czy istnieje potrzeba wydzielenia biura dla osoby pracującej umysłowo, warsztat itp.',
			type: 'long_text'
		},
		{
			id: 'p5',
			number: 5,
			text: 'Czy są przedmioty, meble lub dekoracje z poprzedniego mieszkania, które chcecie przenieść? Jeśli tak to opisz je.',
			type: 'long_text'
		},
		{
			id: 'p6',
			number: 6,
			text: 'Czy istnieje możliwość zmiany położenia przyłączy hydraulicznych, punktów elektrycznych?',
			type: 'long_text'
		},
		{
			id: 'p7',
			number: 7,
			text: 'Jakie są ulubione kolory domowników?',
			type: 'long_text'
		},
		{
			id: 'p8',
			number: 8,
			text: 'Czy są jakieś kolory, materiały, które absolutnie nie powinny się znaleźć we wnętrzu?',
			type: 'long_text'
		},
		{
			id: 'p9',
			number: 9,
			text: 'Jakie są wasze preferencje co do wykończenia podłogi w pokojach?',
			type: 'radio',
			options: [
				{ label: 'Panele winylowe', value: 'panele_winylowe' },
				{ label: 'Podłoga drewniana', value: 'podloga_drewniana' },
				{ label: 'Panele laminowane', value: 'panele_laminowane' }
			]
		},
		{
			id: 'p10',
			number: 10,
			text: 'Jaką szerokość powinna mieć zmywarka w kuchni?',
			type: 'radio',
			options: [
				{ label: '45 cm', value: '45cm' },
				{ label: '60 cm', value: '60cm' }
			]
		},
		{
			id: 'p11',
			number: 11,
			text: 'Jaka lodówka w kuchni? (wolnostojąca, do zabudowy, podwójna, wysoka, niska, z opcją robienia lodu itd.)',
			type: 'long_text'
		},
		{
			id: 'p12',
			number: 12,
			text: 'Jaki dodatkowy sprzęt będzie znajdował się w kuchni? (mikrofala, thermomix, ekspres do kawy itd.)',
			type: 'long_text'
		},
		{
			id: 'p13',
			number: 13,
			text: 'Czy w kuchni ma się znajdować wyspa?',
			type: 'radio',
			options: [
				{ label: 'Tak', value: 'tak' },
				{ label: 'Nie', value: 'nie' },
				{ label: 'Zależy od układu', value: 'zalezy_od_ukladu' }
			]
		},
		{
			id: 'p14',
			number: 14,
			text: 'Jaka kuchenka powinna znaleźć się w kuchni?',
			type: 'radio',
			options: [
				{ label: 'Elektryczna', value: 'elektryczna' },
				{ label: 'Gazowa', value: 'gazowa' },
				{ label: 'Indukcja', value: 'indukcja' }
			]
		},
		{
			id: 'p15',
			number: 15,
			text: 'Jakie minimalne wymiary powinno mieć łóżko w sypialni głównej?',
			type: 'long_text'
		},
		{
			id: 'p16',
			number: 16,
			text: 'Zaznacz właściwe sanitariaty, które mają być uwzględnione przy projekcie łazienki (o ile to możliwe):',
			type: 'multi_select',
			options: [
				{ label: 'Wanna', value: 'wanna' },
				{ label: 'Kabina prysznicowa', value: 'kabina_prysznicowa' },
				{ label: 'Pisuar', value: 'pisuar' },
				{ label: 'Bidet', value: 'bidet' },
				{ label: 'Miska WC', value: 'miska_wc' },
				{ label: 'Jedna umywalka', value: 'jedna_umywalka' },
				{ label: 'Dwie małe umywalki', value: 'dwie_male_umywalki' },
				{ label: 'Pralka', value: 'pralka' },
				{ label: 'Suszarka', value: 'suszarka' }
			]
		},
		{
			id: 'p17',
			number: 17,
			text: 'Czy należy przewidzieć miejsce na ładowanie szczoteczek elektrycznych / golarki / irygatora?',
			type: 'long_text'
		},
		{
			id: 'p18',
			number: 18,
			text: 'Opisz rodzaj ogrzewania zastosowany w projektowanych pomieszczeniach (grzejniki, ogrzewanie podłogowe, ogrzewanie kanałowe):',
			type: 'text'
		},
		{
			id: 'p19',
			number: 19,
			text: 'Czy jest możliwa ingerencja w konstrukcję (przesunięcia, wyburzenia ścianek działowych, powiększenie otworów drzwiowych, wnęk itp.)?',
			type: 'long_text'
		},
		{
			id: 'p20',
			number: 20,
			text: 'Jakie są inne sprzęty codziennego użytku, które powinny być uwzględnione w projekcie?',
			type: 'long_text'
		},
		{
			id: 'p21',
			number: 21,
			text: 'Preferuję styl (zaznacz maksymalnie 3 odpowiedzi):',
			type: 'multi_select',
			maxSelections: 3,
			allowOther: true,
			options: [
				{ label: 'Japandi', value: 'japandi' },
				{ label: 'Nowoczesny', value: 'nowoczesny' },
				{ label: 'Wabi-sabi', value: 'wabi_sabi' },
				{ label: 'Prowansalski', value: 'prowansalski' },
				{ label: 'Glamour', value: 'glamour' },
				{ label: 'Minimalistyczny', value: 'minimalistyczny' },
				{ label: 'Skandynawski', value: 'skandynawski' },
				{ label: 'Angielski', value: 'angielski' },
				{ label: 'Mid-century modern', value: 'mid_century_modern' },
				{ label: 'PRL / Vintage', value: 'prl_vintage' },
				{ label: 'Francuski', value: 'francuski' }
			]
		},
		{
			id: 'p22',
			number: 22,
			text: 'Czy posiadacie zwierzęta? Jeśli tak to jakie i jakie elementy wyposażenia należy przewidzieć? (drapak, kuweta, miejsce na miski itd.)',
			type: 'long_text'
		},
		{
			id: 'p23',
			number: 23,
			text: 'Czy uprawiacie sport ogółem lub w domu? Jeśli tak, to na jakie sprzęty należy przewidzieć miejsce (deska snowboardowa, maty do ćwiczeń, hantle, rower stacjonarny itd.)',
			type: 'long_text'
		},
		{
			id: 'p24',
			number: 24,
			text: 'Czy posiadacie instrumenty muzyczne? Jeśli tak to jakie?',
			type: 'long_text'
		},
		{
			id: 'p25',
			number: 25,
			text: 'Czy często przyjmujecie gości? Czy sofa powinna mieć funkcję rozkładania?',
			type: 'long_text'
		},
		{
			id: 'p26',
			number: 26,
			text: 'Czy należy przewidzieć miejsce na odkurzacz centralny?',
			type: 'long_text'
		},
		{
			id: 'p27',
			number: 27,
			text: 'Czy w projekcie powinien być przewidziany system inteligentnego domu? (np. sterowanie oświetleniem, ogrzewaniem, klimatyzacją)',
			type: 'long_text'
		},
		{
			id: 'p29',
			number: 28,
			text: 'Wasze sugestie, specjalne potrzeby.',
			type: 'long_text'
		}
	],

	thankYou: {
		title: 'Dziękujemy!',
		text: 'Dziękujemy za wypełnienie ankiety! Skontaktujemy się z Tobą wkrótce.'
	},

	gdprText: 'Administratorem danych jest Secco Studio. Dane przetwarzamy w celu przygotowania projektu (art. 6 ust. 1 lit. b RODO).',
	privacyPolicyUrl: '/polityka-prywatnosci'
};
