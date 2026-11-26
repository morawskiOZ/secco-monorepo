# Treści strony

Wszystko, co widać w sekcjach **Projekty**, **Treści** i **Proces projektowy**,
leży w tym katalogu jako zwykłe pliki. Najwygodniej edytować je w panelu
`/admin/` (Sveltia CMS); można też zmieniać pliki ręcznie.

```
projects/<nazwa>.md          jeden plik = jeden projekt (nazwa pliku = adres strony)
articles/<nazwa>.md          jeden plik = jeden artykuł
pages/proces-projektowy.md   strona „Proces projektowy”
media/                       zdjęcia (oryginały, najlepiej do 2560 px)
```

## Szkic czy opublikowane

Każdy wpis ma pole **Szkic** (`draft: true`). Szkice są widoczne tylko w wersji
testowej (stage2). Wpis bez tego znacznika pojawi się na stronie głównej, gdy
sekcje treści zostaną tam włączone. Dzięki temu treści można zbierać i oglądać
już teraz.

Pliki zaczynające się od `przyklad-` to wpisy testowe — do usunięcia, gdy
pojawią się prawdziwe.

## Co przygotować dla jednego projektu

- [ ] Tytuł (np. „Mieszkanie na Krzykach”) i 1–2 zdania opisu na listę projektów
- [ ] Miejsce, powierzchnia, rok — opcjonalnie
- [ ] Zdjęcie główne (poziome wygląda najlepiej na liście) i opis tego, co na nim widać
- [ ] 5–20 zdjęć do galerii, każde z krótkim opisem (pomieszczenie, co jest na zdjęciu)
- [ ] Kilka akapitów: kim są inwestorzy, czego potrzebowali, co zaproponowało studio, jakie materiały
- [ ] Zgoda właścicieli na publikację zdjęć wnętrza
- [ ] Kolejność na liście (1 = pierwszy)

## Co przygotować dla artykułu

- [ ] Tytuł i 1–2 zdania zajawki
- [ ] Treść z śródtytułami (300–1000 słów)
- [ ] Zdjęcie główne i ewentualne zdjęcia w treści
- [ ] 1–3 tematy (np. Porady, Kuchnia, Małe mieszkania)

## Strona „Proces projektowy”

Jest szkic z sześcioma typowymi etapami. Do sprawdzenia i poprawienia przez
studio: czy etapy i ich kolejność zgadzają się z rzeczywistą współpracą, ile
jest tur poprawek, orientacyjne terminy, co klient dostaje na koniec. Warto
dodać po jednym zdjęciu lub rysunku do etapu.

## Zdjęcia

- Panel sam zmniejsza zdjęcia do 2560 px i zapisuje je jako WebP.
- Strona sama przygotowuje mniejsze wersje na telefony — nie trzeba nic przycinać.
- Nazwy plików bez spacji i polskich znaków (panel robi to automatycznie).
- Opis zdjęcia („alt”) jest czytany przez czytniki ekranu i wyszukiwarki — warto go wpisać.
