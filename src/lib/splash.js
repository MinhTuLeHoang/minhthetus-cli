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
  226, 220, 214,                                  // Yellow/Gold
  118, 112, 106, 82, 76, 70, 46, 40, 34,          // Green
  33, 39, 45, 51                                  // Blue -> Cyan Loop
];

function getGradientColor(index, total) {
  const colorIndex = Math.floor((index / total) * gradientColors.length);
  return `\x1b[38;5;${gradientColors[Math.min(colorIndex, gradientColors.length - 1)]}m`;
}

async function showSplash() {
  const icon = "✦";
  const suffix = "+";
  const text = "MINH THE TUS CLI";
  
  // Use standard ASCII but spaced out for premium look
  const styledText = text.toUpperCase().split('').join(' ');
  
  let frame = 0;
  process.stdout.write("\x1b[?25l"); // Hide cursor

  const drawLogo = (isFirst = false) => {
    // \r moves the cursor back to the start of the current line
    // \x1b[K clears the rest of the line (in case the new text is shorter)
    process.stdout.write("\r\x1b[K  ");
    
    let output = `\x1b[38;5;51m${colors.bright}${icon}${colors.reset}   `;
    
    // Gradient text
    for (let i = 0; i < styledText.length; i++) {
       const color = getGradientColor(i + frame, styledText.length + 10);
       output += `${color}${colors.bright}${styledText[i]}${colors.reset}`;
    }
    
    output += `   \x1b[38;5;51m${colors.bright}${suffix}${colors.reset}`;
    process.stdout.write(output);
  };

  // Animate for 5 seconds
  return new Promise((resolve) => {
    const interval = setInterval(() => {
       if (frame < 100) {
          drawLogo();
          frame++;
       } else {
          clearInterval(interval);
          process.stdout.write("\n\n\x1b[?25h"); // Final newlines and show cursor
          resolve();
       }
    }, 40);
  });
}

module.exports = { showSplash };
