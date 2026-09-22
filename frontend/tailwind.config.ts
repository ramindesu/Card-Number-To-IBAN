import type { Config } from 'tailwindcss'

export default <Partial<Config>>{
  theme: {
    extend: {
      colors: {
        ink: '#10251f',
        mint: '#e8fbf3',
        emerald: {
          50: '#effcf6',
          100: '#d8f8e8',
          500: '#18b77a',
          600: '#0b9662',
          700: '#08794f',
        },
        gold: '#d9a441',
      },
      boxShadow: {
        soft: '0 24px 80px rgba(16, 37, 31, 0.12)',
        card: '0 12px 35px rgba(10, 70, 52, 0.12)',
      },
      fontFamily: {
        sans: ['Vazirmatn', 'IRANSans', 'Tahoma', 'sans-serif'],
      },
    },
  },
}
