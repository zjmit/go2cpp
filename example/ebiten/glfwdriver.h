// SPDX-License-Identifier: Apache-2.0

#ifndef GLFWGAME_H
#define GLFWGAME_H

#include "autogen/game.h"

struct GLFWwindow;

class GLFWDriver : public go2cpp_autogen::Game::Driver {
public:
  bool Init() override;
  void Update(std::function<void()> f) override;
  int GetScreenWidth() override;
  int GetScreenHeight() override;
  double GetDevicePixelRatio() override;
  void* GetOpenGLFunction(const char* name) override;
  int GetTouchCount() override;
  void GetTouch(int index, int* id, int* x, int* y) override;

private:
  GLFWwindow* window_;
  double device_pixel_ratio_;
};

#endif
