import { createSignal } from 'solid-js';
import {Button} from "@/components/ui/button";

export default function Home() {
  const [count, setCount] = createSignal(0);

  return (
    <section class="bg-gray-100 text-gray-700 p-8">
      <h1 class="text-2xl font-bold">Home</h1>
      <p class="mt-4">This is the home page.</p>

      <div class="flex items-center space-x-2">
        <Button
          type="button"
          class="border rounded-lg px-2 border-gray-900"
          onClick={() => setCount(count() - 1)}
        >
          -
        </Button>

        <output class="p-10px">Count: {count()}</output>

        <Button
          type="button"
          class="border rounded-lg px-2 border-gray-900"
          onClick={() => setCount(count() + 1)}
        >
          +
        </Button>
      </div>
    </section>
  );
}
