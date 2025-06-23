import { Suspense, type Component } from 'solid-js';
import { A, useLocation } from '@solidjs/router';
import {Navbar} from "@/navigation/navbar";

const App: Component = (props: { children: Element }) => {
  const location = useLocation();

  return (
    <>
      <Navbar/>

      <main>
        <Suspense>{props.children}</Suspense>
      </main>
    </>
  );
};

export default App;
