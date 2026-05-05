/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			colors: {
				'win95':      '#C0C0C0',
				'win-dark':   '#808080',
				'win-navy':   '#000080',
				'win-accent': '#0000FF',
				'win-red':    '#FF0000',
				'win-yellow': '#FFFF00',
				'win-green':  '#00FF00',
				'win-panel':  '#FFFFCC',
			},
			fontFamily: {
				system:  ['"MS Sans Serif"', '"Segoe UI"', 'Tahoma', 'Geneva', 'Verdana', 'sans-serif'],
				heading: ['"Arial Black"', 'Impact', 'Haettenschweiler', 'sans-serif'],
				mono:    ['"Courier New"', 'Courier', 'monospace'],
			},
			borderRadius: {
				DEFAULT: '0px',
				none:    '0px',
				sm:      '0px',
				md:      '0px',
				lg:      '0px',
				xl:      '0px',
				'2xl':   '0px',
				'3xl':   '0px',
				full:    '0px',
			},
		}
	},
	plugins: []
};
