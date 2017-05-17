#version 330
uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

in vec3 vert;
in vec3 vertNormal;
in vec2 vertTexCoord;

out vec3 vertCameraPos;
out vec3 fragNormal;
out vec2 fragTexCoord;

void main() {
    vec4 vertPos4 = model * camera * vec4(vert, 1.0);
    vertCameraPos = vec3(vertPos4) / vertPos4.w;

    fragNormal = vertNormal;
    fragTexCoord = vertTexCoord;
    gl_Position = projection * camera * model * vec4(vert, 1);
}
