export type QuestionType = 'text' | 'long_text' | 'radio' | 'multi_select' | 'file_upload' | 'contact_form';

export interface Option {
	label: string;
	value: string;
}

export interface ContactField {
	id: string;
	label: string;
	type: 'text' | 'email' | 'tel';
	required?: boolean;
	placeholder?: string;
}

export interface Question {
	id: string;
	number: number;
	text: string;
	subtitle?: string;
	type: QuestionType;
	required?: boolean;
	options?: Option[];
	allowOther?: boolean;
	maxSelections?: number;
	maxFileSize?: number;
	maxFiles?: number;
	fields?: ContactField[];
}

export interface FormConfig {
	welcome: { title: string; text: string };
	questions: Question[];
	thankYou: { title: string; text: string };
	gdprText: string;
	privacyPolicyUrl: string;
}

export type FormAnswers = Record<string, string | string[] | File[]>;

export const formConfig: FormConfig = {
	welcome: {
		title: 'Hej!',
		text: 'Hej, miło mi Cię poznać! Odpowiedz na kilka prostych pytań, a ja przygotuję dla Ciebie wycenę projektu.'
	},

	questions: [
		{
			id: 'q1',
			number: 1,
			text: 'Jak trafiłeś na tą ankietę?',
			type: 'radio',
			allowOther: true,
			options: [
				{ label: 'Grupy wnętrzarskie na Facebooku', value: 'facebook' },
				{ label: 'Instagram', value: 'instagram' },
				{ label: 'Sama mi ją wysłałaś :)', value: 'direct' }
			]
		},
		{
			id: 'q2',
			number: 2,
			text: 'Jaki jest metraż powierzchni projektowej?',
			type: 'text'
		},
		{
			id: 'q3',
			number: 3,
			text: 'Czy wcześniej wspomniana powierzchnia dotyczy całej inwestycji czy poszczególnych pomieszczeń (wymień jakich)?',
			type: 'long_text'
		},
		{
			id: 'q4',
			number: 4,
			text: 'Gdzie zlokalizowana jest inwestycja?',
			subtitle: '(Miasto i dzielnica)',
			type: 'text'
		},
		{
			id: 'q5',
			number: 5,
			text: 'Czy mieszkanie/dom jest w stanie deweloperskim czy pochodzi z rynku wtórnego?',
			type: 'text'
		},
		{
			id: 'q6',
			number: 6,
			text: 'Jeśli mieszkanie/dom pochodzi z rynku pierwotnego, na kiedy planowane jest oddanie inwestycji?',
			type: 'text'
		},
		{
			id: 'q7',
			number: 7,
			text: 'Czy posiadasz rzut zakupionej inwestycji? Jeśli tak, to dołącz plik do formularza.',
			type: 'file_upload',
			maxFileSize: 10 * 1024 * 1024,
			maxFiles: 1
		},
		{
			id: 'q8',
			number: 8,
			text: 'Na jakim piętrze znajduję się mieszkanie?',
			type: 'long_text'
		},
		{
			id: 'q9',
			number: 9,
			text: 'Czy wśród projektowanych pomieszczeń, znajdują się takie ze skosami?',
			type: 'radio',
			options: [
				{ label: 'Tak', value: 'yes' },
				{ label: 'Nie', value: 'no' }
			]
		},
		{
			id: 'q10',
			number: 10,
			text: 'Jaki styl wnętrzarski najbardziej Ci się podoba?',
			type: 'multi_select',
			allowOther: true,
			options: [
				{ label: 'Japandi', value: 'japandi' },
				{ label: 'Glamour', value: 'glamour' },
				{ label: 'Skandynawski', value: 'skandynawski' },
				{ label: 'Boho', value: 'boho' },
				{ label: 'Modern classic', value: 'modern_classic' },
				{ label: 'Nowoczesny', value: 'nowoczesny' },
				{ label: 'Industrialny', value: 'industrialny' },
				{ label: 'Klasyczny', value: 'klasyczny' },
				{ label: 'Rustykalny', value: 'rustykalny' },
				{ label: 'Vintage', value: 'vintage' }
			]
		},
		{
			id: 'q11',
			number: 11,
			text: 'Czy potrzebujesz projektu w pełnym zakresie (projekt funkcjonalny, wizualizacje oraz rysunki techniczne wraz z listą zakupową)?',
			type: 'long_text'
		},
		{
			id: 'q12',
			number: 12,
			text: 'Czy jesteś zainteresowany nadzorem autorskim? A może chcesz zamówić pojedyncze wizyty, na których zweryfikuje zgodność wykonanych prac z projektem?',
			type: 'long_text'
		},
		{
			id: 'q13',
			number: 13,
			text: 'Kiedy chciałbyś/chciałabyś rozpocząć prace nad projektem?',
			type: 'radio',
			options: [
				{ label: 'Najszybciej jak jest to możliwe', value: 'asap' },
				{ label: 'W ciągu 1-3 miesięcy', value: '1-3_months' },
				{ label: 'W ciągu 4-7 miesięcy', value: '4-7_months' },
				{ label: 'Nie zależy mi na czasie, jeśli oferta spełni moje oczekiwania jestem w stanie poczekać :)', value: 'no_rush' }
			]
		},
		{
			id: 'q14',
			number: 14,
			text: 'Podaj mi swoje dane adresowe, a ja wrócę do Ciebie ze wstępną wyceną najszybciej jak to będzie możliwe :)',
			type: 'contact_form',
			required: true,
			fields: [
				{ id: 'first_name', label: 'Imię', type: 'text', required: true, placeholder: 'Jan' },
				{ id: 'last_name', label: 'Nazwisko', type: 'text', placeholder: 'Kowalski' },
				{ id: 'phone', label: 'Numer telefonu', type: 'tel', placeholder: '+48 600 000 000' },
				{ id: 'email', label: 'Email', type: 'email', required: true, placeholder: 'jan@example.com' }
			]
		}
	],

	thankYou: {
		title: 'Dziękuję!',
		text: 'Dziękuję! Wrócę do Ciebie ze wstępną wyceną najszybciej jak to będzie możliwe.'
	},

	gdprText: 'Administratorem danych jest Secco Studio. Dane przetwarzamy w celu przygotowania wyceny (art. 6 ust. 1 lit. b RODO).',
	privacyPolicyUrl: '/polityka-prywatnosci'
};
