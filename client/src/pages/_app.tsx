import '../app/globals.css'
import type { AppProps } from "next/app";
import SelectedIssuerProvider from '@/providers/SelectedIssuerProvider';

function MyApp({Component, pageProps}: AppProps) {
  return (
    <SelectedIssuerProvider>
      <Component {...pageProps} />
    </SelectedIssuerProvider>
  );
}

export default MyApp;