// @ts-nocheck — plain browser script, inlined as text (not bundled or type-checked).
// Full-screen viewer for project photos. Loaded as an inline script on project
// pages; without JavaScript the photo links simply open the image file.
(() => {
	const links = [...document.querySelectorAll('a[data-zoom]')];
	const dialog = document.querySelector('dialog.lightbox');
	if (!links.length || !dialog) return;
	const img = dialog.querySelector('img');
	const counter = dialog.querySelector('.lightbox-count');
	let at = 0;

	const show = (i) => {
		at = (i + links.length) % links.length;
		img.src = links[at].href;
		img.alt = links[at].querySelector('img')?.alt ?? '';
		counter.textContent = `${at + 1} / ${links.length}`;
	};

	links.forEach((link, i) =>
		link.addEventListener('click', (e) => {
			if (e.ctrlKey || e.metaKey || e.shiftKey) return;
			e.preventDefault();
			show(i);
			dialog.showModal();
		})
	);
	dialog.querySelector('.lightbox-prev').addEventListener('click', () => show(at - 1));
	dialog.querySelector('.lightbox-next').addEventListener('click', () => show(at + 1));
	dialog.querySelector('.lightbox-close').addEventListener('click', () => dialog.close());
	// A click on the backdrop (the dialog itself, not its content) closes it.
	dialog.addEventListener('click', (e) => e.target === dialog && dialog.close());
	dialog.addEventListener('keydown', (e) => {
		if (e.key === 'ArrowLeft') show(at - 1);
		if (e.key === 'ArrowRight') show(at + 1);
	});
	let startX = null;
	dialog.addEventListener('touchstart', (e) => (startX = e.touches[0].clientX), { passive: true });
	dialog.addEventListener('touchend', (e) => {
		if (startX === null) return;
		const dx = e.changedTouches[0].clientX - startX;
		if (Math.abs(dx) > 50) show(at + (dx < 0 ? 1 : -1));
		startX = null;
	});
})();
