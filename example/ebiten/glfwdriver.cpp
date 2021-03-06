// SPDX-License-Identifier: Apache-2.0

#include "glfwdriver.h"

#include <GLFW/glfw3.h>
#include <dlfcn.h>

namespace {

constexpr int kWidth = 640;
constexpr int kHeight = 480;

}

bool GLFWDriver::Init() {
  if (!glfwInit()) {
    return false;
  }

  glfwWindowHint(GLFW_CLIENT_API, GLFW_OPENGL_API);
  glfwWindowHint(GLFW_CONTEXT_VERSION_MAJOR, 2);
  glfwWindowHint(GLFW_CONTEXT_VERSION_MINOR, 1);

  // TODO: Close the window at the destructor.
  window_ = glfwCreateWindow(kWidth, kHeight, "Ebiten test", nullptr, nullptr);
  if (!window_) {
    glfwTerminate();
    return false;
  }
  glfwMakeContextCurrent(window_);
  glfwSwapInterval(1);

  int framebuffer_width;
  glfwGetFramebufferSize(window_, &framebuffer_width, nullptr);
  device_pixel_ratio_ = static_cast<double>(framebuffer_width) / kWidth;

  return true;
}

void GLFWDriver::Update(std::function<void()> f) {
  glfwPollEvents();
  f();
  glfwSwapBuffers(window_);
}

int GLFWDriver::GetScreenWidth() {
  return kWidth;
}

int GLFWDriver::GetScreenHeight() {
  return kHeight;
}

double GLFWDriver::GetDevicePixelRatio() {
  return device_pixel_ratio_;
}

void* GLFWDriver::GetOpenGLFunction(const char* name) {
  return dlsym(RTLD_DEFAULT, name);
}

int GLFWDriver::GetTouchCount() {
  if (glfwGetMouseButton(window_, GLFW_MOUSE_BUTTON_LEFT) != GLFW_PRESS) {
    return 0;
  }
  return 1;
}

void GLFWDriver::GetTouch(int idx, int* id, int* x, int* y) {
  if (glfwGetMouseButton(window_, GLFW_MOUSE_BUTTON_LEFT) != GLFW_PRESS) {
    // Not reached!
    assert(false);
    return;
  }

  if (id) {
    *id = 0;
  }

  double xpos, ypos;
  glfwGetCursorPos(window_, &xpos, &ypos);
  if (x) {
    *x = static_cast<int>(xpos);
  }
  if (y) {
    *y = static_cast<int>(ypos);
  }
}
