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

let linesBelow = 0;

function trackLines(count = 1) {
  linesBelow += count;
}

function resetLines() {
  linesBelow = 0;
}

async function showSplash(wait = true) {
  const icon = "✦";
  const suffix = "✦";
  const text = "MINH THE TUS CLI";
  
  const styledText = text.toUpperCase().split('').join(' ');
  
  let frame = 0;
  process.stdout.write("\x1b[?25l"); // Hide cursor

  const drawLogo = (final = false) => {
    // If we have lines below, jump up to the logo line
    if (linesBelow > 0) {
      process.stdout.write(`\x1b[${linesBelow}A`);
    }

    process.stdout.write("\r\x1b[K  ");

    const wrapFrame = final ? 0 : frame;
    const iconColor = `\x1b[38;5;${gradientColors[wrapFrame % gradientColors.length]}m`;
    let output = `${iconColor}${colors.bright}${icon}${colors.reset}   `;
    
    const multiplier = 1.3;
    for (let i = 0; i < styledText.length; i++) {
       const color = getGradientColor(i * multiplier + wrapFrame);
       output += `${color}${colors.bright}${styledText[i]}${colors.reset}`;
    }
    
    const suffixColor = `\x1b[38;5;${gradientColors[(wrapFrame + 20) % gradientColors.length]}m`;
    output += `   ${suffixColor}${colors.bright}${suffix}${colors.reset}`;
    process.stdout.write(output);

    // If we jumped up, jump back down to restore cursor position
    if (linesBelow > 0) {
      process.stdout.write(`\x1b[${linesBelow}B\r`);
    }
  };

  const targetFrames = 96;
  
  const animation = new Promise((resolve) => {
    const interval = setInterval(() => {
       if (frame < targetFrames) {
          drawLogo();
          frame++;
       } else {
          clearInterval(interval);
          drawLogo(true);
          process.stdout.write("\x1b[?25h");
          resolve();
       }
    }, 160);
  });

  if (wait) {
    process.stdout.write("\n\n\n\n\n");
    linesBelow = 2; // In wait mode, we just stay on the logo line mostly or finish there
    await animation;
  } else {
    // 2 lines top + 1 splash line + 2 lines bottom = 5 newlines
    process.stdout.write("\n\n\n\n\n");
    linesBelow = 3; // Account for the 2 bottom lines + current line to jump back to splash line
    return animation;
  }
}

module.exports = { showSplash, trackLines, resetLines };
