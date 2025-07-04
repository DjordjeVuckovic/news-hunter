
export const Navbar = () => {
    return (
        <>
            <nav class="bg-gray-800 p-4">
                <div class="container mx-auto flex justify-between items-center">
                    <a href="/" class="text-white text-lg font-bold">News Hunter</a>
                    <div class="space-x-4">
                        <a href="/" class="text-gray-300 hover:text-white">Home</a>
                        <a href="/about" class="text-gray-300 hover:text-white">About</a>
                        <a href="/contact" class="text-gray-300 hover:text-white">Contact</a>
                    </div>
                </div>
            </nav>
        </>
    )
}