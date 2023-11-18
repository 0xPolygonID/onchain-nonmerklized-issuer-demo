import '../app/globals.css'
import type { AppProps } from "next/app";
import { ThemeProvider } from "theme-ui";
import {Component} from "react"

function MyApp({Component, pageProps}: AppProps) {
  return (<Component {...pageProps} />);
}

export default MyApp;