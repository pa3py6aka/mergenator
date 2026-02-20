if (location.pathname.endsWith('mergenator')) {
    const mergenator = await import("./mergenator.js");
    mergenator.start();
}
if (location.pathname.endsWith('tools')) {
    const tools = await import("./tools.js");
    tools.start();
}
