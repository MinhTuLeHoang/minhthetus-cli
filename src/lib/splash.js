const colors = {
  reset: "\x1b[0m",
  bright: "\x1b[1m",
  dim: "\x1b[2m",
};

const gradientColors = [
  51, 50, 49, 45, 44, 43, 39, 38, 37,             // Cyan/Aqua
  33, 32, 31, 27, 26, 25, 21, 20, 19,             // Deep Blue
  57, 56, 55, 93, 92, 91, 129, 128, 127,          // Purple/Indigo
  165, 164, 163, 201, 199, 198, 197,              // Pink/Magenta
  196, 160, 124,                                  // Red
  208, 202, 166,                                  // Orange
  226, 220, 214,                                  // Yellow/Gold
  118, 112, 106, 82, 76, 70, 46, 40, 34,          // Green
  33, 39, 45, 51                                  // Blue -> Cyan Loop
];

function getGradientColor(index) {
  const colorIndex = Math.floor(index) % gradientColors.length;
  return `\x1b[38;5;${gradientColors[colorIndex]}m`;
}

async function showSplash() {
  const icon = "✦";
  const suffix = "✦";
  const text = "MINH THE TUS CLI";
  
  // Use standard ASCII but spaced out for premium look
  const styledText = text.toUpperCase().split('').join(' ');
  
  let frame = 0;
  process.stdout.write("\x1b[?25l"); // Hide cursor

  const drawLogo = (isFirst = false, final = false) => {
    // \r moves the cursor back to the start of the current line
    // \x1b[K clears the rest of the line (in case the new text is shorter)
    process.stdout.write("\r\x1b[K  ");

    const wrapFrame = final ? 0 : frame;
    
    // Icon (Cyan at end, or pulsing)
    const iconColor = `\x1b[38;5;${gradientColors[wrapFrame % gradientColors.length]}m`;
    let output = `${iconColor}${colors.bright}${icon}${colors.reset}   `;
    
    // Gradient text
    // We use a multiplier (1.8) to "squish" the gradient so the full spectrum fits
    const multiplier = 1.3;
    for (let i = 0; i < styledText.length; i++) {
       const color = getGradientColor(i * multiplier + wrapFrame);
       output += `${color}${colors.bright}${styledText[i]}${colors.reset}`;
    }
    
    const suffixColor = `\x1b[38;5;${gradientColors[(wrapFrame + 20) % gradientColors.length]}m`;
    output += `   ${suffixColor}${colors.bright}${suffix}${colors.reset}`;
    process.stdout.write(output);
  };

  // Animate for a specific number of frames to land on the desired colors
  // Frame 48 approx matches the requested screenshot colors with 1.8x scale
  const targetFrames = 96;
  
  return new Promise((resolve) => {
    const interval = setInterval(() => {
       if (frame < targetFrames) {
          drawLogo();
          frame++;
       } else {
          clearInterval(interval);
          drawLogo(false, true); // Final draw with fixed icon colors
          process.stdout.write("\n\n\x1b[?25h"); // Final newlines and show cursor
          resolve();
       }
    }, 40);
  });
}

module.exports = { showSplash };
