#version 330
uniform sampler2D tex;
uniform vec3 ambient;
uniform vec3 sunDir;
uniform vec3 sunDiffuse;
uniform vec3 sunSpecular;
uniform float shininess;

in vec3 fragNormal;
in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
    outputColor = texture(tex, fragTexCoord) * vec4(ambient, 1.0);
}
