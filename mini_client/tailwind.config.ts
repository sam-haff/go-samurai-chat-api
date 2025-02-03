//import type { Config } from "tailwindcss";

/*export default {
  content: ["./src/**///*.{html,js,svelte,ts}"],
/*
  theme: {
    extend: {}
  },

  plugins: [require("@tailwindcss/typography")]
} as Config;*/

import type { Config } from 'tailwindcss';
import flowbitePlugin from 'flowbite/plugin'

export default {
  content: ['./src/**/*.{html,js,svelte,ts}', './node_modules/flowbite-svelte/**/*.{html,js,svelte,ts}'],
  darkMode: 'selector',
  theme: {
    extend: {
      colors: {
        // flowbite-svelte
        /*primary: {
          50: '#FFF5F2',
          100: '#FFF1EE',
          200: '#FFE4DE',
          300: '#FFD5CC',
          400: '#FFBCAD',
          500: '#FE795D',
          600: '#EF562F',
          700: '#EB4F27',
          800: '#CC4522',
          900: '#A5371B'
        },*/
        primary: {
            50: '#effaff',
            100: '#def4ff',
            200: '#b6ebff',
            300: '#75deff',
            400: '#2ccfff',
            500: '#00bfff',
            550: '#00a3e3',
            600: '#0095d4',
            700: '#0076ab',
            750: '#016794',
            800: '#00638d',
            900: '#065374',
            950: '#04344d',
    },
    
      }
    }
  },
  plugins: [flowbitePlugin]
} as Config;
